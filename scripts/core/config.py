import boto3
import os
from typing import Optional

_session: Optional[boto3.Session] = None

def get_aws_session() -> boto3.Session:
    """
    Returns a singleton boto3 Session object.
    Ensures region is always resolved.
    """
    global _session
    if _session is None:
        region = (
            os.environ.get("AWS_REGION")
            or os.environ.get("AWS_DEFAULT_REGION")
            or boto3.Session().region_name
        )
        if not region:
            raise RuntimeError("AWS region could not be determined. Please set AWS_REGION or AWS_DEFAULT_REGION.")
        _session = boto3.session.Session(region_name=region)
    return _session

def get_aws_region() -> str:
    """
    Returns the AWS region resolved from session or env.
    """
    session = get_aws_session()
    region = session.region_name
    if not region:
        raise RuntimeError("AWS region could not be determined.")
    return region
