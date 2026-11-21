from decimal import Decimal
from pydantic import BaseModel, Field


class Cost(BaseModel):
    input_cost_usd: Decimal = Field(ge=0)
    output_cost_usd: Decimal = Field(ge=0)
    cached_input_cost_usd: Decimal | None = Field(default=None, ge=0)
    total_input_cost_usd: Decimal
    reasoning_cost_usd: Decimal | None = Field(default=None, ge=0)
    only_output_cost_usd: Decimal = Field(ge=0)
    total_cost_usd: Decimal = Field(ge=0)
