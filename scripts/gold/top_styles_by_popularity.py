import logging
from typing import List
from pyspark.sql import functions as f

from core.control import ProcessingControl
from .processor import BaseProcessor

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)

class Processor(BaseProcessor):
    """
    Processor for generating the 'top_styles_by_popularity' table in the gold layer.

    This processor reads review_flat data from the silver layer, aggregates
    review counts and average ratings by beer style and month, and writes the result.
    """

    def __init__(self, spark, controls: List[ProcessingControl]):
        super().__init__(spark, controls)

    def run(self):
        try:
            df = self.read()

            result = (
                df.groupBy(
                    "beer_style",
                    f.year(f.from_unixtime("review_time").cast("timestamp")).alias("review_year"),
                    f.month(f.from_unixtime("review_time").cast("timestamp")).alias("review_month"),
                )
                .agg(
                    f.count("*").alias("total_reviews"),
                    f.avg("review_overall").alias("avg_rating"),
                )
                .filter(f.col("total_reviews") >= 25)
                .orderBy(f.col("avg_rating").desc())
            )

            self.write(result)

            logger.info(f"[{self.table}] Processing completed successfully")

        except Exception as e:
            logger.exception(f"[{self.table}] Failed to process top styles by popularity")
            raise
