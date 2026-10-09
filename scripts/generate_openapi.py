"""Build the checked-in API reference. Run after contract changes."""
import json
from pathlib import Path

paths = {}
def ref(name): return {"$ref": "#/components/schemas/" + name}
def response(description, schema=None):
    result = {"description": description}
    if schema:
        result["content"] = {"application/json": {"schema": schema}}
    return result

def operation(path, method, summary, status, model=None, body=None, security=None, errors=()):
    entry = {"summary": summary, "responses": {str(status): response(summary, ref(model) if model else None)}}
    if security: entry["security"] = [{security: []}]
    if body: entry["requestBody"] = {"required": True, "content": {"application/json": {"schema": ref(body)}}}
    for code in errors: entry["responses"][str(code)] = response("Request rejected", ref("Error"))
    if "{chirpID}" in path:
        entry["parameters"] = [{"name": "chirpID", "in": "path", "required": True, "schema": {"type": "string", "format": "uuid"}}]
    paths.setdefault(path, {})[method] = entry

operation("/api/users", "post", "Create an account", 201, "User", "Credentials", errors=(400,409,413,500))
operation("/api/users", "put", "Update the authenticated account", 200, "User", "Credentials", "AccessToken", (400,401,409,413,500))
operation("/api/login", "post", "Login and issue access and refresh tokens", 200, "LoginResponse", "Credentials", errors=(400,401,413,500))
operation("/api/refresh", "post", "Issue a one-hour access token", 200, "AccessResponse", security="RefreshToken", errors=(401,500))
operation("/api/revoke", "post", "Revoke a refresh token", 204, security="RefreshToken", errors=(401,500))
operation("/api/chirps", "post", "Publish as the authenticated user", 201, "Chirp", "ChirpInput", "AccessToken", (400,401,413,500))
operation("/api/chirps", "get", "List a page of chirps", 200, errors=(400,500))
paths["/api/chirps"]["get"]["responses"]["200"] = response("Chirp array", {"type":"array", "items":ref("Chirp")})
paths["/api/chirps"]["get"]["parameters"] = [{"name":name,"in":"query","schema":schema} for name,schema in [
    ("author_id",{"type":"string","format":"uuid"}),
    ("sort",{"type":"string","enum":["asc","desc"],"default":"asc"}),
    ("limit",{"type":"integer","minimum":1,"maximum":100,"default":100}),
    ("offset",{"type":"integer","minimum":0,"maximum":2147483647,"default":0})]]
operation("/api/chirps/{chirpID}","get","Read a chirp",200,"Chirp",errors=(400,404,500))
operation("/api/chirps/{chirpID}","delete","Delete an owned chirp",204,security="AccessToken",errors=(400,401,403,404,500))
operation("/api/polka/webhooks","post","Process a user.upgraded event; ignore other events",204,body="Webhook",security="PolkaKey",errors=(400,401,404,500))
operation("/api/healthz","get","Process readiness",200)
paths["/api/healthz"]["get"]["responses"]["200"]={"description":"Ready","content":{"text/plain":{"schema":{"type":"string","example":"OK"}}}}
operation("/admin/metrics","get","Static request count",200)
paths["/admin/metrics"]["get"]["responses"]["200"]={"description":"HTML metrics","content":{"text/html":{"schema":{"type":"string"}}}}
operation("/admin/reset","post","Delete all users in dev only; reset request count",200,errors=(403,500))
uid={"type":"string","format":"uuid"}
time={"type":"string","format":"date-time"}
user={"type":"object","required":["id","created_at","updated_at","email","is_chirpy_red"],"properties":{"id":uid,"created_at":time,"updated_at":time,"email":{"type":"string","format":"email"},"is_chirpy_red":{"type":"boolean"}}}
login={"allOf":[ref("User"),{"type":"object","required":["token","refresh_token"],"properties":{"token":{"type":"string"},"refresh_token":{"type":"string","pattern":"^[a-f0-9]{64}$"}}}]}
schemas={
 "User":user,"LoginResponse":login,
 "Credentials":{"type":"object","required":["email","password"],"properties":{"email":{"type":"string","format":"email"},"password":{"type":"string","minLength":1,"maxLength":1024,"writeOnly":True}}},
 "Chirp":{"type":"object","required":["id","created_at","updated_at","body","user_id"],"properties":{"id":uid,"created_at":time,"updated_at":time,"body":{"type":"string","maxLength":140},"user_id":uid}},
 "ChirpInput":{"type":"object","required":["body"],"properties":{"body":{"type":"string","minLength":1,"maxLength":140}}},
 "AccessResponse":{"type":"object","required":["token"],"properties":{"token":{"type":"string"}}},
 "Error":{"type":"object","required":["error"],"properties":{"error":{"type":"string"}}},
 "Webhook":{"type":"object","required":["event"],"properties":{"event":{"type":"string","example":"user.upgraded"},"data":{"type":"object","properties":{"user_id":uid}}}},
}
doc={"openapi":"3.0.3","info":{"title":"Threadline Social Publishing API","version":"1.0.0","description":"Accounts, authenticated short posts, refresh sessions, and membership webhooks. Compatibility names from Chirpy are retained."},"servers":[{"url":"http://localhost:8080"}],"paths":paths,"components":{"schemas":schemas,"securitySchemes":{"AccessToken":{"type":"http","scheme":"bearer","bearerFormat":"JWT"},"RefreshToken":{"type":"http","scheme":"bearer","description":"Opaque refresh token"},"PolkaKey":{"type":"apiKey","in":"header","name":"Authorization","description":"ApiKey YOUR_POLKA_KEY"}}}}
Path("docs/openapi.json").write_text(json.dumps(doc,indent=2)+"\n")
