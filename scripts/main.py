import argparse
import logging
import sys

from pyspark.sql import SparkSession
from core.cfn import get_stack_output
from core.enums import LayerType
from core.executor import execute

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

def parse_args():
    """
    Parses command-line arguments for ETL layer, table, and optional parameters.

    Returns:
        argparse.Namespace: Parsed arguments including:
            --layer (required): The data lake layer to process (e.g., silver, gold).
            --table (required): The table to process.
    """
    parser = argparse.ArgumentParser(description="DataMaster ETL Pipeline")
    parser.add_argument(
        "--layer",
        required=True,
        choices=[LayerType.SILVER, LayerType.GOLD],
        help=f"Which layer to process ({LayerType.SILVER} or {LayerType.GOLD})"
    )
    parser.add_argument(
        "--table",
        required=True,
        help="Table name to process"
    )

    try:
        return parser.parse_args()
    except SystemExit as e:
        logger.error(f"Failed to parse command-line arguments: {e}")
        logger.error("Argparse triggered SystemExit (probably invalid args). sys.argv = %s", sys.argv, exc_info=True)
        raise

def main():
    args = None

    try:
        args = parse_args()
        datalake_path = f"s3://{get_stack_output('dm-storage', 'DataLakeBucketName')}"

        spark = (
            SparkSession.builder.appName(f"dm-jobs-{args.layer}")
            .config("spark.sql.extensions", "org.apache.iceberg.spark.extensions.IcebergSparkSessionExtensions")
            .config("spark.sql.catalog.glue_catalog", "org.apache.iceberg.spark.SparkCatalog")
            .config("spark.sql.catalog.glue_catalog.catalog-impl", "org.apache.iceberg.aws.glue.GlueCatalog")
            .config("spark.sql.catalog.glue_catalog.warehouse", datalake_path)
            .config("spark.sql.catalog.glue_catalog.io-impl", "org.apache.iceberg.aws.s3.S3FileIO")
            .config("spark.sql.defaultCatalog", "glue_catalog")
            .config("spark.sql.parquet.compression.codec", "snappy")
            .getOrCreate()
        )

        if args.layer in [LayerType.SILVER, LayerType.GOLD]:
            execute(spark, args.layer, args.table)
        else:
            raise ValueError(f"Unsupported layer: {args.layer}")

    except Exception as e:
        layer = args.layer if args else "unknown"
        table = args.table if args else "unknown"
        logger.error(f"ETL failed for table '{table}' in layer '{layer}': {e}")
        sys.exit(1)  # status != 0 ensures Step Function detects failure

if __name__ == "__main__":
    main()
