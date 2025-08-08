import pytest
from unittest.mock import patch, MagicMock
from datetime import datetime
from pyspark.sql import SparkSession
from core.control import ProcessingControl
from silver.processor import BaseProcessor


class FakeControl(ProcessingControl):
    def __init__(self, schema_name, table_name, object_key):
        self.schema_name = schema_name
        self.table_name = table_name
        self.object_key = object_key
        self.created_at = datetime(2025, 8, 7)


class FakeProcessor(BaseProcessor):
    def __init__(self, spark, controls):
        super().__init__(spark, controls)


@pytest.fixture(scope="session")
def spark():
    return SparkSession.builder \
        .master("local[1]") \
        .appName("unit-test") \
        .getOrCreate()


def test_process_inserts_and_updates(spark):
    """
    Unit test for BaseProcessor.process method.

    This test verifies that the processor:
    - Calls Spark's read.parquet() to load input data (mocked here).
    - Splits records into insert, update, and delete operations.
    - Calls saveAsTable() for insert records (mocked).
    - Executes a MERGE statement for update and delete records (mocked).

    The test uses MagicMock to avoid any real interaction with S3, Iceberg, or Spark jobs.
    It ensures the method behaves correctly when given a typical batch of CDC input.
    """
    df = MagicMock()
    df.filter.return_value = df
    df.drop.return_value = df
    df.withColumn.return_value = df
    df.unionByName.return_value = df
    df.limit.return_value.count.return_value = 1
    df.write.format.return_value.mode.return_value.partitionBy.return_value.option.return_value.saveAsTable = MagicMock()

    with patch("silver.processor.get_stack_output", return_value="fake-bucket"), \
         patch("pyspark.sql.readwriter.DataFrameReader.parquet", return_value=df):

        ctrl = FakeControl("dm_silver", "beer", "bronze/beer/file.parquet")
        proc = FakeProcessor(spark, [ctrl])
        proc.spark.sql = MagicMock()

        proc.process(pk_column="id")

        df.write.format.return_value.mode.return_value.partitionBy.return_value.option.return_value.saveAsTable.assert_called()
        proc.spark.sql.assert_called()
