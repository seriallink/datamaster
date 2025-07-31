from enum import Enum

class LayerType(str, Enum):
    SILVER = "silver"
    GOLD = "gold"

class ProcessingStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    ERROR = "error"
