import json

# It seems that Ansible has a bug and/or j2 doesn't work without a
# global function
def from_json(a, *args, **kw):
    ''' Convert the value to JSON '''
    return json.loads(a)
