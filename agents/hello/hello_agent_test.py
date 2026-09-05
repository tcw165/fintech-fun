from hello_agent import (
    DEFAULT_INSTRUCTION,
    DEFAULT_MODEL,
    DEFAULT_NAME,
    build_hello_agent,
)


def test_build_hello_agent_sets_identity():
    agent = build_hello_agent()
    assert agent.name == DEFAULT_NAME
    assert agent.model == DEFAULT_MODEL
    assert agent.instruction == DEFAULT_INSTRUCTION


def test_build_hello_agent_accepts_overrides():
    agent = build_hello_agent(
        name="custom_agent",
        model="gemini-2.0-flash",
        instruction="Stay brief.",
    )
    assert agent.name == "custom_agent"
    assert agent.model == "gemini-2.0-flash"
    assert agent.instruction == "Stay brief."
