"""Give every operation in the Deutero spec a runnable example.

Two rules, deliberately asymmetric:

* READ operations use real ids from the operator's account, so the Phase 5 live
  matrix exercises real data rather than 404ing on placeholders.
* DESTRUCTIVE operations (DELETE, and secret rotation) always use obviously fake
  ids, so no example — run by a human, an agent, or `dogfood --allow-destructive`
  — can ever delete or rotate a real resource.
"""
import json
import sys

SPEC = sys.argv[1]

REAL = {
    "study_id": "f0ed9214-06fc-4420-87b3-31240b670fc1",
    "project_id": "be8fc6bf-8819-4f73-81f9-961cf751b13a",
    "interview_id": "caeaec18-cab8-44fb-9be4-fd45dac9efe4",
    "webhook_id": "9ca6a913-54ff-495d-92f0-2baaa24ee9c6",
    "simulation_id": "4cf77793-9e1c-4c2f-afbf-958fd0c60305",
    "persona_id": "5f8e4e84-aa61-42f5-81ed-67cd6afdd87a",
    "question_id": "425202fa-68bf-46b6-babb-b323c0f10629",
}
# Recognisably fake, UUID-shaped so validation passes.
FAKE = {name: "00000000-0000-4000-8000-%012d" % i for i, name in enumerate(
    ["study_id", "project_id", "interview_id", "webhook_id", "simulation_id",
     "persona_id", "question_id", "translation_id", "key_id"], start=1)}
FAKE_DEFAULT = "00000000-0000-4000-8000-000000000099"

QUERY_BY_NAME = {
    "q": "pricing",
    "category": "text",
    "external_participant_id": "panel-42",
    "question_id": REAL["question_id"],
    "question_number": 1,
    "event_type": "interview.completed",
}

s = json.load(open(SPEC))
schemas = s.get("components", {}).get("schemas", {})


def resolve(schema):
    ref = (schema or {}).get("$ref")
    if ref:
        return schemas.get(ref.split("/")[-1], {})
    return schema or {}


def synth(prop_name, schema):
    schema = resolve(schema)
    for key in ("anyOf", "oneOf"):
        if key in schema:
            for alt in schema[key]:
                alt = resolve(alt)
                if alt.get("type") != "null":
                    return synth(prop_name, alt)
    if "example" in schema:
        return schema["example"]
    if schema.get("enum"):
        return schema["enum"][0]
    if "default" in schema and schema["default"] is not None:
        return schema["default"]
    t = schema.get("type")
    if t == "integer":
        return 1
    if t == "number":
        return 1.0
    if t == "boolean":
        return True
    if t == "array":
        item = synth(prop_name, schema.get("items", {}))
        return [item] if item is not None else []
    if t == "object":
        return {k: synth(k, v) for k, v in (schema.get("properties") or {}).items()
                if k in (schema.get("required") or [])}
    # strings, by field name
    n = prop_name.lower()
    if n.endswith("_id") and n in REAL:
        return REAL[n]
    if "url" in n:
        return "https://example.com/deutero-hook"
    if "email" in n:
        return "research@example.com"
    if n in ("name", "label", "title"):
        return "Example (created by deutero-pp-cli docs)"
    if n in ("text", "question", "prompt", "description", "body", "content"):
        return "How did you first hear about us?"
    if n in ("type", "question_type"):
        return "text"
    if n in ("language", "locale"):
        return "es"
    return "example"


def is_destructive(method, path):
    return method == "delete" or "rotate" in path


counts = {"path": 0, "query": 0, "body": 0, "defanged": 0}
for path, item in s["paths"].items():
    for method, op in item.items():
        if method not in ("get", "post", "put", "patch", "delete"):
            continue
        destructive = is_destructive(method, path)
        for prm in op.get("parameters") or []:
            name = prm.get("name")
            if prm.get("in") == "path":
                if destructive:
                    fake = FAKE.get(name, FAKE_DEFAULT)
                    if prm.get("example") != fake:
                        prm["example"] = fake
                        counts["defanged"] += 1
                elif "example" not in prm:
                    prm["example"] = REAL.get(name, FAKE.get(name, FAKE_DEFAULT))
                    counts["path"] += 1
            elif prm.get("in") == "query" and prm.get("required") and "example" not in prm:
                prm["example"] = QUERY_BY_NAME.get(name, synth(name, prm.get("schema", {})))
                counts["query"] += 1
        rb = (op.get("requestBody") or {}).get("content", {}).get("application/json")
        if rb is not None and "example" not in rb:
            body_schema = resolve(rb.get("schema", {}))
            example = {k: synth(k, v) for k, v in (body_schema.get("properties") or {}).items()
                       if k in (body_schema.get("required") or [])}
            # "question_id or question_number is required" is enforced in code,
            # not in the schema, so the synthesiser cannot see it.
            if path.endswith("/analysis/optimal-clusters") or path.endswith("/analysis/cluster"):
                example.setdefault("question_number", 1)
            if example:
                rb["example"] = example
                counts["body"] += 1

json.dump(s, open(SPEC, "w"), indent=2)
print("examples added:", counts)

# ---- second pass: property-level examples on every request-body schema -------
# The generator builds a body command's Example from per-property examples on the
# body schema, not from a media-level example, so seed those too.
body_schema_names = set()
for path, item in s["paths"].items():
    for method, op in item.items():
        rb = (op.get("requestBody") or {}).get("content", {}).get("application/json") if isinstance(op, dict) else None
        if rb:
            ref = (rb.get("schema") or {}).get("$ref")
            if ref:
                body_schema_names.add(ref.split("/")[-1])
prop_count = 0
for name in body_schema_names:
    sch = schemas.get(name, {})
    for pname, pschema in (sch.get("properties") or {}).items():
        if "example" in pschema:
            continue
        val = synth(pname, pschema)
        if val is None or isinstance(val, (dict, list)) and not val:
            continue
        pschema["example"] = val
        prop_count += 1
json.dump(s, open(SPEC, "w"), indent=2)
print("body-schema property examples added:", prop_count, "across", len(body_schema_names), "schemas")
