
try:
    import sys as _sys
    from jinja2.filters import pass_context as _passctx, pass_environment as _passenv, pass_eval_context as _passevalctx
    _mod = _sys.modules['jinja2.filters']
    _mod.contextfilter = _passctx
    _mod.environmentfilter = _passenv
    _mod.evalcontextfilter = _passevalctx
except ImportError:
    _sys = None

from ansible.plugins.filter.core import *
