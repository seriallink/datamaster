import logging
import importlib
from pyspark.sql import SparkSession
from core.dynamo import fetch_pending_controls

logger = logging.getLogger(__name__)

def process(spark: SparkSession, table: str):
    """
    Executes the silver-layer transformation for a specific table.

    Steps:
    - Fetches all 'pending' ProcessingControl entries for the given table from DynamoDB.
    - Dynamically loads the processor module for the table.
    - Iterates over each control individually:
        - Marks the status as "running".
        - Executes the table-specific processing logic.
        - Updates status as "success" or "error".

    Raises:
        Exception: If module loading or processing fails for any control, the error is logged and re-raised.
    """
    try:
        controls = fetch_pending_controls("dm_silver", table)
        logger.info(f"Found {len(controls)} pending files for table {table}")
    except Exception as e:
        logger.exception("Failed to fetch pending items from DynamoDB")
        raise

    if not controls:
        logger.info(f"No pending files found for table: {table}")
        return

    try:
        module = importlib.import_module(f"silver.{table}")
        processor_class = getattr(module, "Processor")
    except ModuleNotFoundError as e:
        logger.warning(f"No processor module found for table: {table}")
        raise e
    except AttributeError as e:
        logger.warning(f"Module 'silver.{table}' does not define 'Processor'")
        raise e

    for ctrl in controls:
        try:
            ctrl.start()
            processor = processor_class(spark, [ctrl])
            processor.run()
            ctrl.finish()
        except Exception as e:
            logger.exception(f"Failed to process control for {ctrl.object_key}")
            ctrl.finish(failure=e)
            raise e
