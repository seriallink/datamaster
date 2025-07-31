import logging
from abc import ABC
from pyspark.sql import SparkSession, Window
from pyspark.sql import functions as f
from typing import List

from core.cfn import get_stack_output
from core.control import ProcessingControl
from core.exceptions import ParquetReadError, DataFrameOperationError, SparkWriteError

logger = logging.getLogger(__name__)

class BaseProcessor(ABC):
    """
    Abstract base class for processors working on a batch of files for a single table.
    Handles reading CDC data from bronze layer, enriching it with timestamps,
    and applying operations (insert, update, soft delete) into Iceberg.
    """

    def __init__(self, spark: SparkSession, controls: List[ProcessingControl]):
        """
        Initializes the processor with a Spark session and a list of ProcessingControl items.

        Args:
            spark (SparkSession): Active Spark session
            controls (List[ProcessingControl]): List of control items for the same table
        """
        if not controls:
            raise ValueError("Processor must be initialized with at least one ProcessingControl")

        self.spark = spark
        self.controls = controls
        self.catalog = "glue_catalog"
        self.schema = controls[0].schema_name
        self.table = controls[0].table_name
        self.layer = controls[0].schema_name.replace("dm_", "")
        self.bucket = get_stack_output("dm-storage", "DataLakeBucketName").rstrip("/")

    def get_parquet_paths(self) -> List[str]:
        """
        Builds full S3 paths for all files referenced in the controls.

        Returns:
            List[str]: Full paths to Parquet files in S3
        """
        return [f"s3://{self.bucket}/{ctrl.object_key}" for ctrl in self.controls]

    def process(self, pk_column: str):
        """
        Executes the CDC pipeline logic for a batch of Parquet files from the bronze layer.

        Steps:
        - Reads all files associated with this batch
        - Enriches each record with created_at / updated_at / deleted_at based on ProcessingControl
        - Adds partitioned_at column for efficient partitioning (derived from updated_at)
        - Splits records by operation: insert, update, delete
        - Applies inserts via append
        - Applies updates and deletes via Iceberg MERGE (soft delete)

        Args:
            pk_column (str): Primary key column used for record matching during update/delete
        """

        def safe_union(dfs):
            if not dfs:
                return None
            result = dfs[0]
            for d in dfs[1:]:
                result = result.unionByName(d)
            return result

        def finalize(d):
            return d.withColumn("partitioned_at", f.date_format("updated_at", "yyyyMMdd")).drop("_control_ts")

        def has_data(d):
            try:
                return d is not None and d.limit(1).count() > 0
            except:
                return False

        try:
            inserts = []
            updates = []
            deletes = []

            for ctrl in self.controls:
                path = f"s3://{self.bucket}/{ctrl.object_key}"
                try:
                    df = self.spark.read.parquet(path)
                except Exception as e:
                    logger.exception(f"[{self.table}] Failed to read file: {path}")
                    raise ParquetReadError(f"Could not read parquet file: {path}") from e

                df = df.withColumn("_control_ts", f.lit(ctrl.created_at.isoformat()))

                inserts.append(
                    df.filter(df.operation == "insert").drop("operation")
                    .withColumn("created_at", f.col("_control_ts").cast("timestamp"))
                    .withColumn("updated_at", f.col("_control_ts").cast("timestamp"))
                    .withColumn("deleted_at", f.lit(None).cast("timestamp"))
                )
                updates.append(
                    df.filter(df.operation == "update").drop("operation")
                    .withColumn("updated_at", f.col("_control_ts").cast("timestamp"))
                    .withColumn("deleted_at", f.lit(None).cast("timestamp"))
                )
                deletes.append(
                    df.filter(df.operation == "delete").drop("operation")
                    .withColumn("updated_at", f.col("_control_ts").cast("timestamp"))
                    .withColumn("deleted_at", f.col("_control_ts").cast("timestamp"))
                )

        except Exception as e:
            logger.exception(f"[{self.schema}.{self.table}] Failed to read or transform input files")
            raise DataFrameOperationError("Failed to process input files") from e

        try:
            df_inserts = safe_union(inserts)
            df_updates = safe_union(updates)
            df_deletes = safe_union(deletes)

            merged = df_updates.unionByName(df_deletes) if df_updates and df_deletes else df_updates or df_deletes

            if df_inserts is not None:
                df_inserts = finalize(df_inserts)

            if merged is not None:
                window = Window.partitionBy(pk_column).orderBy(f.col("updated_at").desc())
                merged = merged.withColumn("rn", f.row_number().over(window)).filter(f.col("rn") == 1).drop("rn")
                merged = finalize(merged)

            logger.info(f"[{self.table}] inserts={df_inserts.count() if df_inserts else 0}, "
                        f"updates={df_updates.count() if df_updates else 0}, "
                        f"deletes={df_deletes.count() if df_deletes else 0}")

        except Exception as e:
            logger.exception(f"[{self.schema}.{self.table}] Failed during dataframe preparation")
            raise DataFrameOperationError("Failed to prepare dataframe") from e

        try:
            if has_data(df_inserts):
                df_inserts.write \
                    .format("iceberg") \
                    .mode("append") \
                    .partitionBy("partitioned_at") \
                    .option("path", f"s3://{self.bucket}/{self.layer}/{self.table}") \
                    .saveAsTable(f"{self.catalog}.{self.schema}.{self.table}")

            if has_data(merged):
                merged.createOrReplaceTempView("merged")
                self.spark.sql(f"""
                    MERGE INTO {self.catalog}.{self.schema}.{self.table} t
                    USING merged s
                    ON t.{pk_column} = s.{pk_column}
                    WHEN MATCHED THEN UPDATE SET *
                    WHEN NOT MATCHED THEN INSERT *
                """)
        except Exception as e:
            logger.exception(f"[{self.schema}.{self.table}] Failed during Spark write operation")
            raise SparkWriteError("Failed during Spark write operation") from e
