import logging
from typing import List
from uuid import uuid4

from core.control import ProcessingControl
from core.exceptions import ProcessingControlPersistError, InvalidObjectKeyFormat

logger = logging.getLogger(__name__)

GOLD_TABLES = [
    "top_beers_by_rating",
    "top_breweries_by_rating",
    "state_by_review_volume",
    "top_styles_by_popularity",
    "top_drinkers"
]

def spawn_controls(source_controls: List[ProcessingControl]):
    """
    Generates and persists new ProcessingControl records for the gold layer,
    based on the successful output of the 'review_flat' table.

    For each source ProcessingControl, this function creates one new control record
    per gold table, reusing the same partition identifier (derived from object_key),
    compute_target, file_format, and record_count.
    """
    for src in source_controls:
        try:
            partition = src.object_key.split("partitioned_at=")[-1]
            if not partition:
                raise InvalidObjectKeyFormat(f"Missing partition in object_key: {src.object_key}")
        except Exception as e:
            logger.exception(f"[{src.table_name}] Failed to extract partition from object_key")
            raise InvalidObjectKeyFormat(f"Error parsing partition from {src.object_key}") from e

        for table_name in GOLD_TABLES:
            try:
                gold_control = ProcessingControl(
                    control_id=str(uuid4()),
                    object_key=f"silver/review_flat/partitioned_at={partition}",
                    schema_name="dm_gold",
                    table_name=table_name,
                    compute_target=src.compute_target,
                    file_format=src.file_format,
                    record_count=src.record_count,
                )
                gold_control.persist()
            except ProcessingControlPersistError:
                logger.exception(f"[{table_name}] Failed to persist gold control for partition {partition}")
                raise
            except Exception:
                logger.exception(f"[{table_name}] Unexpected error while creating gold control")
                raise
