from tfx.components import CsvExampleGen
from tfx.dsl.components.base.base_node import BaseNode
from tfx.orchestration.data_types import RuntimeParameter


def create_components() -> list[BaseNode]:
    return [CsvExampleGen(input_base=RuntimeParameter(name="input_base", ptype=str, default="/data"))]
