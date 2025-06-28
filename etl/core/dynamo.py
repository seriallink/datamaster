from typing import List, Optional
from .cfn import get_stack_outputs
from .config import get_aws_session
from .control import ProcessingControl
from .enums import ProcessingStatus
from .exceptions import DynamoQueryError, StackOutputNotFoundError

def fetch_pending_controls(schema_name: str, table_name: Optional[str] = None) -> List[ProcessingControl]:
    """
    Fetches all 'pending' ProcessingControl objects for a given schema (and optional table) from DynamoDB.

    Args:
        schema_name (str): The schema name (e.g., dm_silver or dm_gold)
        table_name (Optional[str]): The table to filter (e.g., 'beer'). If None, returns all pending files in schema.

    Returns:
        List[ProcessingControl]: List of control records with status 'pending'.
    """
    try:
        outputs = get_stack_outputs("dm-control")
        table_name_dynamo = outputs.get("ProcessingControlTableName")
        index_name = outputs.get("SchemaStatusIndexName")
    except Exception as e:
        raise StackOutputNotFoundError("Failed to load required stack outputs") from e

    try:
        client = get_aws_session().client("dynamodb")
        kwargs = {
            "TableName": table_name_dynamo,
            "IndexName": index_name,
            "KeyConditionExpression": "#schema = :s AND #status = :p",
            "ExpressionAttributeNames": {
                "#schema": "schema_name",
                "#status": "status",
            },
            "ExpressionAttributeValues": {
                ":s": {"S": schema_name},
                ":p": {"S": ProcessingStatus.PENDING.value},
            },
        }

        if table_name:
            kwargs["FilterExpression"] = "#table = :t"
            kwargs["ExpressionAttributeNames"]["#table"] = "table_name"
            kwargs["ExpressionAttributeValues"][":t"] = {"S": table_name} # type: ignore

        response = client.query(**kwargs)
        items = response.get("Items", [])
        controls = [ProcessingControl.from_dynamo_item(item) for item in items]
        return controls

    except Exception as e:
        raise DynamoQueryError(f"DynamoDB query failed: {e}") from e
