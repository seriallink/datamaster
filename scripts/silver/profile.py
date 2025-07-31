from typing import List
from core.control import ProcessingControl
from .processor import BaseProcessor

class Processor(BaseProcessor):
    def __init__(self, spark, controls: List[ProcessingControl]):
        super().__init__(spark, controls)

    def run(self):
        self.process("profile_id")
