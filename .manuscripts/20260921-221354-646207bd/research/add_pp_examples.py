"""Add x-pp-example / x-happy-args where the generator cannot synthesise them.

The generator derives a command's Example only from REQUIRED inputs. Body-only
commands (create/update/set) and endpoints whose "one of these is required" rule
lives in code rather than the schema therefore ship with no Example. This reads
the built CLI's own command tree (agent-context) to get the exact command path
and flag names, then writes an explicit example back onto the spec operation.
"""
import json
import subprocess
import sys

SPEC, BINARY = sys.argv[1], sys.argv[2]

REAL = {"study_id": "f0ed9214-06fc-4420-87b3-31240b670fc1",
        "project_id": "be8fc6bf-8819-4f73-81f9-961cf751b13a",
        "interview_id": "caeaec18-cab8-44fb-9be4-fd45dac9efe4",
        "webhook_id": "9ca6a913-54ff-495d-92f0-2baaa24ee9c6",
        "simulation_id": "4cf77793-9e1c-4c2f-afbf-958fd0c60305",
        "persona_id": "5f8e4e84-aa61-42f5-81ed-67cd6afdd87a",
        "question_id": "425202fa-68bf-46b6-babb-b323c0f10629"}
FAKE = "00000000-0000-4000-8000-000000000099"

# Body-field values that make an example meaningful. Anything not listed falls
# back to the schema's own example, which enrich_examples.py already seeded.
FLAG_VALUES = {
    "name": '"Example study"', "label": '"Example webhook"',
    "url": "https://example.com/deutero-hook", "text": '"How did you first hear about us?"',
    "type": "text", "events": "interview.completed", "question-number": "1",
}

ctx = json.loads(subprocess.run([BINARY, "agent-context"], capture_output=True, text=True, check=True).stdout)
spec = json.load(open(SPEC))
schemas = spec.get("components", {}).get("schemas", {})


def walk(cmds, prefix):
    for c in cmds:
        path = prefix + [c["name"]]
        yield path, c
        yield from walk(c.get("subcommands") or [], path)


by_endpoint = {}
for path, c in walk(ctx["commands"], []):
    ann = c.get("annotations") or {}
    if "pp:path" in ann and "pp:method" in ann:
        by_endpoint[(ann["pp:method"].lower(), ann["pp:path"])] = (path, c)


def required_body_fields(op):
    rb = (op.get("requestBody") or {}).get("content", {}).get("application/json", {})
    ref = (rb.get("schema") or {}).get("$ref", "")
    sch = schemas.get(ref.split("/")[-1], {}) if ref else rb.get("schema", {})
    return list(sch.get("required") or []), sch.get("properties") or {}


for _p, _it in spec["paths"].items():
    for _m, _op in _it.items():
        if isinstance(_op, dict):
            _op.pop("x-pp-example", None)

added = []
for api_path, item in spec["paths"].items():
    for method, op in item.items():
        if method not in ("get", "post", "put", "patch", "delete"):
            continue
        hit = by_endpoint.get((method, api_path))
        if not hit:
            continue
        cmd_path, cmd = hit
        needs_selector = api_path.endswith(("/analysis/optimal-clusters", "/analysis/cluster"))
        if not op.get("requestBody") and not needs_selector:
            continue  # the generator already synthesises these from required params
        destructive = method == "delete" or "rotate" in api_path
        # positional args, in the order the command's Use line declares them
        positionals = [tok.strip("<>[]") for tok in cmd.get("use", "").split()[1:] if tok.startswith("<")]
        pos_vals = [(FAKE if destructive else REAL.get(p, FAKE)) for p in positionals]

        req, props = required_body_fields(op)
        flag_tokens = []
        for field in req:
            flag = field.replace("_", "-")
            val = FLAG_VALUES.get(flag)
            if val is None:
                ex = (props.get(field) or {}).get("example")
                if ex is None or isinstance(ex, (dict, list)):
                    continue
                val = str(ex)
                if " " in val:
                    val = '"%s"' % val
            flag_tokens.append("--%s %s" % (flag, val))

        for prm in op.get("parameters") or []:
            if prm.get("in") == "query" and prm.get("required") and prm.get("example") is not None:
                v = str(prm["example"])
                flag_tokens.append("--%s %s" % (prm["name"].replace("_", "-"), '"%s"' % v if " " in v else v))

        # The API enforces "question_id or question_number" in code, so no field
        # is schema-required and the generator emits nothing to select a question.
        happy = None
        if api_path.endswith(("/analysis/optimal-clusters", "/analysis/cluster")):
            flag_tokens.append("--question-number 1")
            happy = "--question-number=1"

        if True:
            parts = ["deutero-pp-cli"] + cmd_path + pos_vals + flag_tokens
            op["x-pp-example"] = "  " + " ".join(parts)
            added.append(" ".join(cmd_path))
        if happy:
            op["x-happy-args"] = happy

json.dump(spec, open(SPEC, "w"), indent=2)
print("x-pp-example added to %d operations" % len(added))
for a in sorted(added)[:40]:
    print("  ", a)
