"""Per-operation example overrides where the generic example hits the wrong fixture.

Run AFTER add_pp_examples.py (which clears x-pp-example before recomputing).

* Scale and options tallies need a scale / options question respectively; the
  generic question id is free-text, so the API correctly answers 400.
* Transcript search's default mode (hybrid) returns 500 in production because of a
  pg8000 paramstyle bug in the trigram query — fixed in the dashboard repo
  (study_api/routers/search.py) but not yet deployed. Semantic mode never reaches
  that query, so the example uses it until the fix ships.
"""
import json, sys
p = sys.argv[1]; s = json.load(open(p))
EA = "f1a767b5-3e58-4f5f-9575-611bf6886735"
over = {
    ("get", "/api/v1/studies/{study_id}/analysis/responses/scale"):
        f"  deutero-pp-cli studies analysis get-scale-responses {EA} --question-id a8c57d33-afa8-4b68-99fb-dd8336442d3f",
    ("get", "/api/v1/studies/{study_id}/analysis/responses/options"):
        f"  deutero-pp-cli studies analysis get-options-responses {EA} --question-id 45127ab0-1869-4113-aaf7-5f3bd3649bf6",
    ("get", "/api/v1/studies/{study_id}/search"):
        "  deutero-pp-cli studies search transcripts f0ed9214-06fc-4420-87b3-31240b670fc1 --q pricing --mode semantic",
}
for (m, path), ex in over.items():
    s["paths"][path][m]["x-pp-example"] = ex

# The live matrix drives happy paths from pp:happy-args (derived from parameter
# examples), not from the Cobra Example, so the fixture must be set there too.
happy = {
    ("get", "/api/v1/studies/{study_id}/analysis/responses/scale"):
        f"study_id={EA};--question-id=a8c57d33-afa8-4b68-99fb-dd8336442d3f",
    ("get", "/api/v1/studies/{study_id}/analysis/responses/options"):
        f"study_id={EA};--question-id=45127ab0-1869-4113-aaf7-5f3bd3649bf6",
}
for (m, path), h in happy.items():
    s["paths"][path][m]["x-happy-args"] = h
json.dump(s, open(p, "w"), indent=2)
print("fixture overrides applied:", len(over))
