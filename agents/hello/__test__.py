import os
import sys

import pytest

if __name__ == "__main__":
    test_path = os.path.dirname(os.path.abspath(__file__))
    sys.exit(pytest.main([test_path, "-v"]))
