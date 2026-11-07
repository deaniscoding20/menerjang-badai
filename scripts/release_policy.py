"""Fail-closed deployment eligibility policy; no credentials in output."""
import argparse
import json
import os
import re
import urllib.request


def authorize(run, pulls, repository, owner):
    sha = run.get("sha", "")
    if not (run.get("event") == "push" and run.get("branch") == "main"
            and run.get("conclusion") == "success"
            and run.get("repository") == repository
            and run.get("status") == "completed"
            and run.get("workflow_path") == ".github/workflows/ci.yml"
            and re.fullmatch(r"[0-9a-f]{40}", sha)):
        return False
    return any(
        pr.get("merged") is True
        and pr.get("merge_commit_sha") == sha
        and pr.get("base", {}).get("ref") == "main"
        and pr.get("base", {}).get("repo", {}).get("full_name") == repository
        and pr.get("merged_by", {}).get("login") == owner
        and pr.get("merged_by", {}).get("id") == 351171974
        for pr in pulls
    )


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--event-file", required=True)
    parser.add_argument("--repository", required=True)
    parser.add_argument("--owner", required=True)
    args = parser.parse_args()
    with open(args.event_file, encoding="utf-8") as stream:
        event = json.load(stream)
    wr = event.get("workflow_run", {})
    if type(wr.get("id")) is not int or wr["id"] <= 0:
        raise SystemExit("Invalid CI run ID")
    headers = {"Authorization": "Bearer " + os.environ["GH_TOKEN"],
               "Accept": "application/vnd.github+json", "User-Agent": "MB_default_agent"}

    def get(path):
        request = urllib.request.Request("https://api.github.com/repos/" + args.repository + path, headers=headers)
        with urllib.request.urlopen(request, timeout=25) as response:
            return json.load(response)

    # Refetch authoritative run metadata, not only the webhook payload.
    wr = get("/actions/runs/" + str(wr["id"]))
    run = {"event": wr.get("event"), "branch": wr.get("head_branch"),
           "conclusion": wr.get("conclusion"), "sha": wr.get("head_sha"),
           "repository": wr.get("head_repository", {}).get("full_name"),
           "status": wr.get("status"), "workflow_path": wr.get("path")}
    if not re.fullmatch(r"[0-9a-f]{40}", run.get("sha") or ""):
        raise SystemExit("Invalid CI commit")
    refs = get("/commits/" + run["sha"] + "/pulls?per_page=100")
    pulls = [get("/pulls/" + str(pr["number"])) for pr in refs]
    eligible = authorize(run, pulls, args.repository, args.owner)
    # A delayed run must not roll back a newer main deployment.
    eligible = eligible and get("/git/ref/heads/main")["object"]["sha"] == run["sha"]
    with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as stream:
        stream.write("eligible=" + str(eligible).lower() + "\n")
        stream.write("sha=" + run["sha"] + "\n")
    print("Owner-merged main release eligible:", eligible)


if __name__ == "__main__":
    main()
