
from decimal import Decimal

from miko.llm.base import Usage
from .cost_model import Cost
from .pricing import PRICING_REGISTRY
from miko.core.logging.logger import get_logger
from miko.core.logging.errors import NotFoundError, InvalidError

okitsune = get_logger("llm.costs")

def calculate_model_total_cost(
    provider: str,
    model: str,
    usage: Usage,
) -> Cost:
    p_key, m_key = provider.strip().lower(), model.strip().lower()
    
    okitsune.debug(
        "Starting price calculation", 
        extra={
            "provider": p_key, 
            "model": m_key
        }
    )
    # ========== Retrieve model pricing ==========
    try:
        price = PRICING_REGISTRY[(p_key, m_key)]
    except KeyError as exc:
        raise NotFoundError(
            "Missing pricing for requested model",
            provider=p_key,
            model=m_key,
            cause=exc
        ) from exc

    m = Decimal("1000000")

    # ========== Validate input tokens ==========
    if usage.cached_tokens > usage.input_tokens:
        raise InvalidError(
            "Cached tokens exceed total input tokens",
            cached_tokens=usage.cached_tokens,
            input_tokens=usage.input_tokens,
        )

    cached_tokens = Decimal(usage.cached_tokens)
    uncached_input_tokens = Decimal(
        usage.input_tokens - usage.cached_tokens
    )

    # ========== Validate output tokens ==========
    if usage.reasoning_tokens > usage.output_tokens:
        raise InvalidError(
            "Reasoning tokens exceed output tokens",
            reasoning_tokens=usage.reasoning_tokens,
            output_tokens=usage.output_tokens
        )

    reasoning_tokens = Decimal(usage.reasoning_tokens)
    only_output_tokens = Decimal(
        usage.output_tokens - usage.reasoning_tokens
    )

    # ========== Validate cached input pricing ==========
    if cached_tokens > 0 and price.cached_input_cost is None:
        raise InvalidError(
            "Missing cached input price",
            provider=p_key,
            model=m_key,
            cached_tokens=usage.cached_tokens,
        )

    cached_cost = (
        price.cached_input_cost
        if price.cached_input_cost is not None
        else Decimal("0")
    )

    # ========== Input price ==========
    input_cost_usd = (
        uncached_input_tokens / m
    ) * price.input_cost

    # ========== Cached input price ==========
    cached_input_usd = (
        cached_tokens / m
    ) * cached_cost

    # ========== Total input price ==========
    total_input_usd = input_cost_usd + cached_input_usd

    # ========== Ordinary output price ==========
    only_output_usd = (
        only_output_tokens / m
    ) * price.output_cost

    # ========== Reasoning price ==========
    reasoning_price = (
        price.reasoning_cost
        if price.reasoning_cost is not None
        else price.output_cost
    )

    reasoning_usd = (
        reasoning_tokens / m
    ) * reasoning_price

    # ========== Total output price ==========
    output_usd = only_output_usd + reasoning_usd

    # ========== Grand total request cost ==========
    total_cost_usd = total_input_usd + output_usd

    okitsune.info(
        "Model pricing calculation succeeded!",
        extra= {
            "provider": p_key,
            "model": m_key,
            "total_cost_usd": str(total_cost_usd),
        },
    )
    return Cost(
        input_cost_usd=input_cost_usd,
        cached_input_cost_usd=cached_input_usd,
        total_input_cost_usd=total_input_usd,
        only_output_cost_usd=only_output_usd,
        reasoning_cost_usd=reasoning_usd,
        output_cost_usd=output_usd,
        total_cost_usd=total_cost_usd,
    )

