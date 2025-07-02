import logging
from abc import ABC
from typing import List
from pyspark.sql import SparkSession, DataFrame
from core.cfn import get_stack_output
from core.control import ProcessingControl
from core.exceptions import ParquetReadError, SparkWriteError

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)

class BaseProcessor(ABC):
    """
    Abstract base class for gold-layer processors.

    Handles loading data from the silver.review_flat table based on the controls,
    and writing the result to an Iceberg table in the gold layer.
    """

    def __init__(self, spark: SparkSession, controls: List[ProcessingControl]):
        """
        Initializes the processor with a Spark session and a list of ProcessingControl items.

        Args:
            spark (SparkSession): Active Spark session
            controls (List[ProcessingControl]): List of controls for the gold table
        """
        if not controls:
            raise ValueError("Processor must be initialized with at least one ProcessingControl")

        self.spark = spark
        self.controls = controls
        self.catalog = "glue_catalog"
        self.schema = controls[0].schema_name
        self.table = controls[0].table_name
        self.bucket = get_stack_output("dm-storage", "DataLakeBucketName").rstrip("/")

    def read(self) -> DataFrame:
        """
        Reads all input review_flat records from the silver layer using the Glue Catalog,
        filtering by the set of partitions present in the controls.

        Returns:
            DataFrame: Union of all relevant review_flat records.
        """
        try:
            partitions = [ctrl.object_key.split("partitioned_at=")[-1] for ctrl in self.controls]
            logger.info(f"[{self.table}] Filtering review_flat with partitions: {partitions}")

            partition_values = ",".join([f"'{p}'" for p in partitions])
            df = self.spark.read \
                .format("iceberg") \
                .table(f"{self.catalog}.dm_silver.review_flat") \
                .filter(f"partitioned_at IN ({partition_values})")

            return df

        except Exception as e:
            logger.exception(f"[{self.table}] Failed to read from Glue table")
            raise ParquetReadError(f"Could not read review_flat from Glue for partitions: {partitions}") from e

    def write(self, df: DataFrame):
        """
        Writes the final result to the Iceberg gold table, partitioned by review_year and review_month.

        Args:
            df (DataFrame): Transformed and aggregated DataFrame to be written

        Raises:
            SparkWriteError: If the write operation fails
        """
        try:
            logger.info(f"[{self.table}] Writing result to {self.schema}.{self.table}")

            df.write \
                .format("iceberg") \
                .mode("append") \
                .partitionBy("review_year", "review_month") \
                .option("path", f"s3://{self.bucket}/gold/{self.table}") \
                .saveAsTable(f"{self.catalog}.{self.schema}.{self.table}")

        except Exception as e:
            logger.exception(f"[{self.table}] Failed to write output data")
            raise SparkWriteError("Failed during Spark write operation") from e

