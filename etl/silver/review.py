from datetime import datetime, timezone
from typing import List
from uuid import uuid4
from core.enums import ProcessingStatus
from core.processor import BaseProcessor, ProcessingControl

class Processor(BaseProcessor):
    def __init__(self, spark, controls: List[ProcessingControl]):
        super().__init__(spark, controls)

    def run(self):
        self.process("review_id")
        self.add_partition_controls()

    def add_partition_controls(self):
        """
        Generates and persists one processing control per unique review partition,
        based on the source controls, to enable incremental processing of review_flat.
        """
        seen_partitions = set()

        for ctrl in self.controls:
            partition = ctrl.created_at.strftime("%Y%m%d")

            if partition in seen_partitions:
                continue

            seen_partitions.add(partition)

            control = ProcessingControl(
                control_id=str(uuid4()),
                object_key=f"silver/review/partitioned_at={partition}",
                schema_name="dm_silver",
                table_name="review_flat",
                record_count=ctrl.record_count,
                status=ProcessingStatus.PENDING,
                compute_target="emr",
                created_at=datetime.now(timezone.utc),
                updated_at=datetime.now(timezone.utc),
            )

            try:
                control.persist()
            except Exception as e:
                raise e
