import logging
from boto3.dynamodb.types import TypeDeserializer, TypeSerializer
from botocore.exceptions import ClientError
from dataclasses import asdict, dataclass, field, fields
from datetime import datetime, timezone
from enum import Enum
from typing import Optional

from .cfn import get_stack_output
from .config import get_aws_session
from .enums import ProcessingStatus
from .exceptions import ProcessingControlPersistError, InvalidObjectKeyFormat

logger = logging.getLogger(__name__)

@dataclass
class ProcessingControl:
    """
    Represents the processing metadata of a single data file throughout the ETL pipeline.
    Mirrors the structure defined in the DynamoDB table 'dm-processing-control'.
    """

    control_id: str
    object_key: str
    schema_name: str
    table_name: str
    record_count: int = 0
    file_format: str = "parquet"
    file_size: Optional[int] = None
    checksum: Optional[str] = None
    status: str = field(default=ProcessingStatus.PENDING)
    attempt_count: int = 0
    compute_target: str = "emr"
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    duration: Optional[int] = None
    error_message: Optional[str] = None
    created_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))
    updated_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))

    @classmethod
    def from_dynamo_item(cls, item: dict) -> "ProcessingControl":
        """
        Deserializes a DynamoDB item (in AttributeValue format) into a ProcessingControl instance.

        Args:
            item (dict): A DynamoDB item with AttributeValue-encoded fields.

        Returns:
            ProcessingControl: The instantiated control object with populated fields.
        """
        deser = TypeDeserializer()
        plain_item = {k: deser.deserialize(v) for k, v in item.items()}

        for f in fields(cls):
            if f.type is datetime and isinstance(plain_item.get(f.name), str):
                try:
                    plain_item[f.name] = datetime.fromisoformat(plain_item[f.name].replace("Z", "+00:00"))
                except Exception:
                    raise

        return cls(**plain_item)

    def destination_key(self, target_layer: str) -> str:
        """
        Returns the destination S3 key for the next layer of the pipeline.
        """
        try:
            parts = self.object_key.split("/")
            filename = parts[-1].replace(".json.gz", "").replace(".csv.gz", "").replace(".gz", "").replace(".parquet", "")
            return f"{target_layer}/{self.table_name}/{filename}.parquet"
        except Exception as e:
            logger.exception(f"Failed to generate destination key from {self.object_key}")
            raise InvalidObjectKeyFormat(f"Cannot derive destination key: {self.object_key}") from e

    def start(self):
        """
        Marks the beginning of processing for this control object.
        """
        self.status = ProcessingStatus.RUNNING
        self.started_at = datetime.now(timezone.utc)
        self.attempt_count += 1
        self.updated_at = datetime.now(timezone.utc)
        logger.info(f"[{self.table_name}] Status set to RUNNING")
        self.persist()

    def finish(self, failure: Optional[Exception] = None):
        """
        Finalizes the processing status, setting success or error and calculating duration.
        """
        self.finished_at = datetime.now(timezone.utc)
        self.updated_at = datetime.now(timezone.utc)

        if self.started_at:
            delta = self.finished_at - self.started_at
            self.duration = int(delta.total_seconds() * 1000)

        if failure:
            self.status = ProcessingStatus.ERROR
            self.error_message = str(failure)[:1000]
            logger.error(f"[{self.table_name}] Finished with ERROR: {self.error_message}")
        else:
            self.status = ProcessingStatus.SUCCESS
            self.error_message = None
            logger.info(f"[{self.table_name}] Finished with SUCCESS")

        self.persist()

    def persist(self):
        """
        Persists this ProcessingControl instance to DynamoDB using automatic serialization.
        """
        try:
            table_name = get_stack_output("dm-control", "ProcessingControlTableName")
        except Exception:
            logger.exception("Failed to resolve CloudFormation stack output")
            raise

        serializer = TypeSerializer()
        plain_item = asdict(self)

        for k, v in plain_item.items():
            if isinstance(v, Enum):
                plain_item[k] = v.value
            elif isinstance(v, datetime):
                plain_item[k] = v.isoformat()

        filtered_item = {k: v for k, v in plain_item.items() if v is not None}
        item = {k: serializer.serialize(v) for k, v in filtered_item.items()}

        try:
            get_aws_session().client("dynamodb").put_item(
                TableName=table_name,
                Item=item
            )
            logger.info(f"[{self.table_name}] Control persisted with status {self.status.value}")
        except ClientError as e:
            logger.exception("DynamoDB put_item failed")
            raise ProcessingControlPersistError("Failed to persist ProcessingControl") from e
