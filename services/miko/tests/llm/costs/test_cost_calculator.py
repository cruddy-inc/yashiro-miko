import pytest
from miko.llm.costs.cost_calculator import calculate_model_total_cost
from miko.llm.base import Usage
from unittest.mock import MagicMock
from miko.core.logging.errors import AppError, InvalidError, NotFoundError
from miko.llm.costs.cost_model import Cost
from decimal import Decimal


@pytest.mark.parametrize(
    ("provider", "model", "usage", "expected_cost"),
    [
        (
            "openai",
            "gpt-6.2-sol",
            Usage(
                input_tokens=30000,
                output_tokens=12000,
                cached_tokens=4580,
            ),
            Cost(
                input_cost_usd=Decimal("0.05084"),
                output_cost_usd=Decimal("0.12"),
                cached_input_cost_usd=Decimal("0.000458"),
                total_input_cost_usd=Decimal("0.051298"),
                reasoning_cost_usd=Decimal("0.00"),
                only_output_cost_usd=Decimal("0.12"),
                total_cost_usd=Decimal("0.171298"),
            ),
        ),
        (
            "anthropic",
            "claude-opus-5-5",
            Usage(
                input_tokens=32000, 
                output_tokens=21040, 
                cached_tokens=6300
            ),
            Cost(
                input_cost_usd=Decimal("0.1028"),
                output_cost_usd=Decimal("0.4208"),
                cached_input_cost_usd=Decimal("0.00126"),
                total_input_cost_usd=Decimal("0.10406"),
                reasoning_cost_usd=Decimal("0.00"),
                only_output_cost_usd=Decimal("0.4208"),
                total_cost_usd=Decimal("0.52486"),
            ),
        ),
        (
            "google-ai",
            "gemini-4-neon",
            Usage(
                input_tokens=42000, 
                output_tokens=19382, 
                cached_tokens=14030,
            ),
            Cost(
                input_cost_usd=Decimal("0.08391"),
                output_cost_usd=Decimal("0.29073"),
                cached_input_cost_usd=Decimal("0.0021045"),
                total_input_cost_usd=Decimal("0.0860145"),
                reasoning_cost_usd=Decimal("0.00"),
                only_output_cost_usd=Decimal("0.29073"),
                total_cost_usd=Decimal("0.3767445"),
            ),
        ),
    ],
)
def test_cost_caclulator_model_mix(
    provider: str,
    model: str,
    usage: Usage,
    expected_cost: Cost,
):
    mock_err = MagicMock(spec=AppError)
    mock_err.message = "Calculation failed miserably"

    result = calculate_model_total_cost(provider=provider, model=model, usage=usage)

    assert result == expected_cost

@pytest.mark.parametrize(
    ("provider", "model", "usage", "expected_err"),
    [
        (
            "openai",
            "gpt-6.2-astra",
            Usage(
                input_tokens=10000,
                output_tokens=12000,
                cached_tokens=15000,
            ),
            InvalidError,
        ),

        (
            "google-ai",
            "gemini-4-carbon",
            Usage(
                input_tokens=12310,
                output_tokens=12000,
                cached_tokens=4032,
                reasoning_tokens=20000,
            ),
            InvalidError,
        ),

        (
            "unknown-provider",
            "non-existent-model",
            Usage(
                input_tokens=1000,
                output_tokens=500,
                cached_tokens=930,
                reasoning_tokens=199
            ),
            NotFoundError,
        ),
    ],
)
def test_calculate_model_total_cost_errors(
    provider: str,
    model: str,
    usage: Usage,
    expected_err: type[Exception],
):
    with pytest.raises(expected_err):
        calculate_model_total_cost(
            provider=provider,
            model=model,
            usage=usage,
        )


def test_calculate_model_total_cost_key_normalization():
    usage = Usage(
        input_tokens=30000,
        output_tokens=12000,
        cached_tokens=4580,
    )

    result_normalized = calculate_model_total_cost(
        provider="  OpEnAI  ",
        model="  Gpt-6.2-SoL  ",
        usage=usage,
    )
    result_standard = calculate_model_total_cost(
        provider="openai",
        model="gpt-6.2-sol",
        usage=usage,
    )

    assert result_normalized == result_standard


def test_calculate_model_total_cost_zero_usage():
    usage = Usage(
        input_tokens=0,
        output_tokens=0,
        cached_tokens=0,
        reasoning_tokens=0,
    )

    result = calculate_model_total_cost(
        provider="openai",
        model="gpt-6.2-sol",
        usage=usage,
    )

    expected = Cost(
        input_cost_usd=Decimal("0"),
        cached_input_cost_usd=Decimal("0"),
        total_input_cost_usd=Decimal("0"),
        only_output_cost_usd=Decimal("0"),
        reasoning_cost_usd=Decimal("0"),
        output_cost_usd=Decimal("0"),
        total_cost_usd=Decimal("0"),
    )

    assert result == expected