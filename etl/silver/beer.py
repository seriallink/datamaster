from typing import List
from core.processor import BaseProcessor, ProcessingControl

class Processor(BaseProcessor):
    def __init__(self, spark, controls: List[ProcessingControl]):
        super().__init__(spark, controls)

    def run(self):
        self.process("beer_id")
