from .llm_price_model import Decimal, LLMPrice

PRICING_REGISTRY: dict[tuple[str, str], LLMPrice] = {
    # =================================================
    # ================= OpenAI models =================
    # =================================================
    ("openai", "gpt-6.2-sol"): LLMPrice(
        input_cost=Decimal("2.00"),
        output_cost=Decimal("10.00"),
        cached_input_cost=Decimal("0.10"),
    ),
    ("openai", "gpt-6.2-astra"): LLMPrice(
        input_cost=Decimal("8.00"),
        output_cost=Decimal("40.00"),
        cached_input_cost=Decimal("0.4"),
    ),
    # =================================================
    # ================ Anthropic models ===============
    # =================================================
    ("anthropic", "claude-sonnet-5-5"): LLMPrice(
        input_cost=Decimal("2.00"),
        output_cost=Decimal("10.00"),
        cached_input_cost=Decimal("0.10"),
    ),
    ("anthropic", "claude-opus-5-5"): LLMPrice(
        input_cost=Decimal("4.00"),
        output_cost=Decimal("20.00"),
        cached_input_cost=Decimal("0.20"),
    ),
    ("anthropic", "claude-fable-5-2"): LLMPrice(
        input_cost=Decimal("8.00"),
        output_cost=Decimal("40.00"),
        cached_input_cost=Decimal("0.20"),
    ),
    # =================================================
    # ================ Google AI models ===============
    # =================================================
    ("google-ai", "gemini-4-carbon"): LLMPrice(
        input_cost=Decimal("4.00"),
        output_cost=Decimal("20.00"),
        cached_input_cost=Decimal("0.20"),
    ),
    ("google-ai", "gemini-4-neon"): LLMPrice(
        input_cost=Decimal("3.00"),
        output_cost=Decimal("15.00"),
        cached_input_cost=Decimal("0.15")
    ), 
}
