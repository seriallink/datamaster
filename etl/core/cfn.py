from typing import Dict
from botocore.exceptions import ClientError, BotoCoreError

from .config import get_aws_session
from .exceptions import StackNotFoundError, StackDescribeError, StackOutputNotFoundError

def get_stack_outputs(stack_name: str) -> Dict[str, str]:
    """
    Retrieves all outputs from a given CloudFormation stack.

    Args:
        stack_name (str): The name of the stack to query.

    Returns:
        Dict[str, str]: A dictionary mapping output keys to their values.

    Raises:
        StackNotFoundError: If the stack does not exist.
        StackDescribeError: If there is a problem retrieving the stack.
    """
    client = get_aws_session().client("cloudformation")

    try:
        response = client.describe_stacks(StackName=stack_name)
    except ClientError as e:
        if e.response["Error"]["Code"] == "ValidationError":
            raise StackNotFoundError(f"No stack found with name '{stack_name}'") from e
        raise StackDescribeError(f"Client error describing stack '{stack_name}': {e}") from e
    except BotoCoreError as e:
        raise StackDescribeError(f"Failed to describe stack '{stack_name}': {e}") from e

    stacks = response.get("Stacks", [])
    if not stacks:
        raise StackNotFoundError(f"No stack found with name '{stack_name}'")

    outputs = stacks[0].get("Outputs", [])
    return {
        o["OutputKey"]: o["OutputValue"]
        for o in outputs
        if "OutputKey" in o and "OutputValue" in o
    }


def get_stack_output(stack_name: str, key: str) -> str:
    """
    Retrieves a specific output value from a CloudFormation stack.

    Args:
        stack_name (str): The name of the stack to query.
        key (str): The output key to retrieve.

    Returns:
        str: The value of the output key.

    Raises:
        StackOutputNotFoundError: If the output key is not found.
        StackNotFoundError / StackDescribeError: If the stack cannot be queried.
    """
    outputs = get_stack_outputs(stack_name)
    if key not in outputs:
        raise StackOutputNotFoundError(f"Output key '{key}' not found in stack '{stack_name}'")
    return outputs[key]
