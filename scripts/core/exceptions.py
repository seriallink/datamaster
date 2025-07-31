class StackNotFoundError(Exception):
    """Raised when the specified CloudFormation stack does not exist."""
    pass

class StackOutputNotFoundError(Exception):
    """Raised when the requested output key is not found in the stack."""
    pass

class StackDescribeError(Exception):
    """Raised when there is an unexpected error describing the stack."""
    pass

class DynamoQueryError(Exception):
    """Raised when a query to DynamoDB fails."""
    pass

class InvalidObjectKeyFormat(Exception):
    """Raised when the S3 object key does not follow the expected structure."""
    pass

class ProcessingControlPersistError(Exception):
    """Raised when persisting the ProcessingControl to DynamoDB fails."""
    pass

class ParquetReadError(Exception):
    """Raised when reading Parquet files fails."""
    pass

class DataFrameOperationError(Exception):
    """Raised when filtering the DataFrame by operation fails."""
    pass

class SparkWriteError(Exception):
    """Raised when a write operation to Iceberg via Spark fails."""
    pass
