import logging
import importlib
from pyspark.sql import SparkSession
from core.dynamo import fetch_pending_controls

logger = logging.getLogger(__name__)

def execute(spark: SparkSession, layer: str, table: str):
    """
    Executes the ETL transformation for a specific table in a given layer.

    Steps:
    - Fetches all 'pending' ProcessingControl entries from DynamoDB.
    - Dynamically loads the processor module for the table in the given layer.
    - Runs the transformation and updates control status.

    Args:
        spark (SparkSession): Active Spark session
        layer (str): Layer name (e.g. 'silver', 'gold')
        table (str): Table name (e.g. 'beer', 'review_flat')

    Raises:
        Exception: if module loading or processing fails
    """
    schema = f"dm_{layer}"

    try:
        controls = fetch_pending_controls(schema, table)
        logger.info(f"[{schema}.{table}] Found {len(controls)} pending files")
    except Exception:
        logger.exception("Failed to fetch pending items from DynamoDB")
        raise

    if not controls:
        logger.info(f"[{schema}.{table}] No pending files found")
        return

    try:
        module = importlib.import_module(f"{layer}.{table}")
        processor_class = getattr(module, "Processor")
    except ModuleNotFoundError as e:
        logger.warning(f"[{layer}.{table}] Module not found")
        raise e
    except AttributeError as e:
        logger.warning(f"[{layer}.{table}] Processor class not defined")
        raise e

    # Group all controls and process in batch
    try:
        for ctrl in controls:
            ctrl.start()

        processor = processor_class(spark, controls)
        processor.run()

        for ctrl in controls:
            ctrl.finish()

    except Exception as e:
        logger.exception(f"[{schema}.{table}] Failed to process control batch")
        for ctrl in controls:
            ctrl.finish(failure=e)
        raise
