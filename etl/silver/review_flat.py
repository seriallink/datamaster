import logging
from pyspark.sql import functions as f
from typing import List
from core.processor import BaseProcessor, ProcessingControl

logger = logging.getLogger(__name__)

class Processor(BaseProcessor):
    def __init__(self, spark, controls: List[ProcessingControl]):
        super().__init__(spark, controls)

    def run(self):
        """
        Builds and writes the 'review_flat' table using control-specified partitions.
        """
        if not self.controls:
            logger.info(f"[{self.table}] No control records to process.")
            return

        partitions = {ctrl.object_key.split("partitioned_at=")[-1] for ctrl in self.controls}
        logger.info(f"[{self.table}] Processing partitions: {sorted(partitions)}")

        table_prefix = f"{self.catalog}.{self.schema}"

        review = (
            self.spark.read.table(f"{table_prefix}.review")
            .filter(f.col("partitioned_at").isin(list(partitions)))
            .alias("r")
        )

        if review.limit(1).count() == 0:
            logger.info(f"[{self.table}] No records to process.")
            return

        beer = self.spark.read.table(f"{table_prefix}.beer").alias("b")
        brewery = self.spark.read.table(f"{table_prefix}.brewery").alias("w")
        profile = self.spark.read.table(f"{table_prefix}.profile").alias("p")

        logger.info(f"[{self.table}] Joining dimensions")

        df = (
            review
            .join(beer, "beer_id", "left")
            .join(brewery, "brewery_id", "left")
            .join(profile, "profile_id", "left")
            .select(
                "r.review_id",
                "r.brewery_id",
                "r.beer_id",
                "r.profile_id",
                "w.brewery_name",
                "b.beer_name",
                "b.beer_style",
                "b.beer_abv",
                "p.profile_name",
                "p.email",
                "p.state",
                "r.review_overall",
                "r.review_aroma",
                "r.review_appearance",
                "r.review_palate",
                "r.review_taste",
                "r.review_time",
                "r.created_at",
                "r.updated_at",
                "r.deleted_at",
                "r.partitioned_at"
            )
        )

        logger.info(f"[{self.schema}.{self.table}] Writing {df.count()} records to Iceberg")

        df.write \
            .format("iceberg") \
            .mode("overwrite") \
            .partitionBy("partitioned_at") \
            .option("path", f"s3://{self.bucket}/{self.layer}/{self.table}") \
            .saveAsTable(f"{self.catalog}.{self.schema}.{self.table}")
