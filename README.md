# fintech-fun

Playground for fin-tech ideas.

## Build

Bazel 8.2 + Python 3.12 (bzlmod). Refresh the pip lock after editing `requirements.in`:

```bash
bazel run //:requirements.update
```

Run the hello Google ADK agent tests (no live Gemini calls):

```bash
bazel test //agents/hello:hello_agent_test
```

