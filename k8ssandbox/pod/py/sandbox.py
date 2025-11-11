from fastapi import FastAPI, Body
import subprocess, tempfile, os

app = FastAPI()

@app.post("/run")
def run_code(code: str = Body(..., embed=True)):
    with tempfile.NamedTemporaryFile(delete=False, suffix=".py") as f:
        f.write(code.encode("utf-8"))
        f.flush()
        try:
            result = subprocess.run(
                ["python3", f.name],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                timeout=3
            )
            return {
                "stdout": result.stdout.decode(),
                "stderr": result.stderr.decode(),
                "returncode": result.returncode
            }
        except subprocess.TimeoutExpired:
            return {"error": "timeout"}

@app.get("/health")
def health():
    return {"status": "idle"}
