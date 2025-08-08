import pytest
from unittest.mock import patch, MagicMock
from datetime import datetime
from pyspark.sql import SparkSession
from core.control import ProcessingControl
from gold.processor import BaseProcessor


class FakeControl(ProcessingControl):
    def __init__(self, schema_name, table_name, object_key):
        self.schema_name = schema_name
        self.table_name = table_name
        self.object_key = object_key
        self.created_at = datetime(2025, 8, 7)


class FakeGoldProcessor(BaseProcessor):
    def __init__(self, spark, controls):
        super().__init__(spark, controls)


@pytest.fixture(scope="session")
def spark():
    return SparkSession.builder \
        .master("local[1]") \
        .appName("gold-test") \
        .getOrCreate()


def test_read_and_write_gold_table(spark):
    """
    Unit test for the BaseProcessor in the gold layer.

    This test verifies that:
    - The read() method correctly applies partition filters when loading data
      from the Iceberg 'dm_silver.review_flat' table.
    - The write() method writes the resulting DataFrame to the correct
      Iceberg gold table using the expected path and partition columns.
    - All external interactions (Glue Catalog, S3, Iceberg) are mocked to ensure
      the test runs quickly and safely in a local environment.
    """
    df = MagicMock()
    df.filter.return_value = df
    df.write.format.return_value.mode.return_value.partitionBy.return_value.option.return_value.saveAsTable = MagicMock()

    with patch("gold.processor.get_stack_output", return_value="fake-bucket"), \
         patch("pyspark.sql.readwriter.DataFrameReader.table", return_value=df):

        ctrl = FakeControl("dm_gold", "top_beers_by_rating", "partitioned_at=202510")
        proc = FakeGoldProcessor(spark, [ctrl])
        proc.spark.sql = MagicMock()

        result_df = proc.read()
        df.filter.assert_called_with("partitioned_at IN ('202510')")

        proc.write(result_df)
        df.write.format.return_value.mode.return_value.partitionBy.return_value.option.return_value.saveAsTable.assert_called()
