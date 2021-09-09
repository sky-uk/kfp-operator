import os
import pytest

from unittest.mock import patch, MagicMock
from tfx.types import channel_utils
from tfx.types import standard_artifacts
from tfx.components import Pusher

from compiler import component_expander

serving_model_dir = "serving_model"

@patch("tfx.components.Pusher", autospec=True)
def test_expand_components_with_pusher_existing(pusher):
    components = component_expander.expand_components_with_pusher([pusher], serving_model_dir)
    assert components == [pusher]


@patch("tfx.components.Trainer", autospec=True)
@patch("tfx.components.Evaluator", autospec=True)
def test_expand_components_with_pusher(evaluator, trainer):
    trainer.outputs = {"model": channel_utils.as_channel([standard_artifacts.Model()])}
    evaluator.outputs = {
        "blessing": channel_utils.as_channel([standard_artifacts.ModelBlessing()])
    }

    components = component_expander.expand_components_with_pusher([trainer, evaluator], serving_model_dir)
    assert trainer in components
    assert evaluator in components
    assert any(isinstance(component, Pusher) for component in components)


@patch("tfx.components.Trainer", autospec=True)
def test_expand_components_with_pusher_two_trainers(trainer):
    trainer.outputs = {"model": channel_utils.as_channel([standard_artifacts.Model()])}

    with pytest.raises(SystemExit) as pytest_wrapped_e:
        component_expander.expand_components_with_pusher([trainer, trainer], serving_model_dir)
    assert pytest_wrapped_e.type == SystemExit

@patch("tfx.components.Trainer", autospec=True)
@patch("tfx.components.Evaluator", autospec=True)
def test_expand_components_with_pusher_two_evaluators(evaluator, trainer):
    trainer.outputs = {"model": channel_utils.as_channel([standard_artifacts.Model()])}
    evaluator.outputs = {
        "blessing": channel_utils.as_channel([standard_artifacts.ModelBlessing()])
    }

    with pytest.raises(SystemExit) as pytest_wrapped_e:
        component_expander.expand_components_with_pusher([trainer, evaluator, evaluator], serving_model_dir)
    assert pytest_wrapped_e.type == SystemExit


def test_expand_components_with_pusher_no_trainer():

    components = component_expander.expand_components_with_pusher([], serving_model_dir)
    assert components == []

@patch("tfx.components.Trainer", autospec=True)
def test_expand_components_with_pusher_no_evaluator(trainer):

    trainer.outputs = {"model": channel_utils.as_channel([standard_artifacts.Model()])}

    components = component_expander.expand_components_with_pusher([trainer], serving_model_dir)
    assert any(isinstance(component, Pusher) for component in components)