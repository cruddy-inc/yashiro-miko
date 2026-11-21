from pydantic import BaseModel, Field
from decimal import Decimal


class LLMPrice(BaseModel):
    input_cost: Decimal = Field(
        ge=0,
        description="Uncached input token cost (USD per 1M tokens)"
    )
    output_cost: Decimal = Field(
        ge=0,
        description="Output/completion token cost (USD per 1M tokens)",
    )
    cached_input_cost: Decimal | None = Field(
        default=None,
        ge=0,
        description="Cached input token read/hit cost (USD per 1M tokens)",
    )
    reasoning_cost: Decimal | None = Field(
        default=None,
        ge=0,
        description="Reasoning token cost, when billed separately (USD per 1M tokens)",
    )