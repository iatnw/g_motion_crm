import assert from "node:assert/strict";
import test from "node:test";
import { auth, AUTH_EXPIRED_EVENT } from "../lib/api.ts";

test("an expired session notifies the UI and rejects the protected request", async (t) => {
  const previousWindow = globalThis.window;
  const window = new EventTarget();
  globalThis.window = window;
  t.after(() => {
    if (previousWindow === undefined) delete globalThis.window;
    else globalThis.window = previousWindow;
  });
  let notifications = 0;
  window.addEventListener(AUTH_EXPIRED_EVENT, () => { notifications++; });
  t.mock.method(globalThis, "fetch", async () => new Response("Unauthorized", { status: 401 }));
  await assert.rejects(auth.me(), /Unauthorized/);
  assert.equal(notifications, 1);
});

test("a rejected password change stays in the form without expiring the session", async (t) => {
  const previousWindow = globalThis.window;
  const window = new EventTarget();
  globalThis.window = window;
  t.after(() => {
    if (previousWindow === undefined) delete globalThis.window;
    else globalThis.window = previousWindow;
  });
  let notifications = 0;
  window.addEventListener(AUTH_EXPIRED_EVENT, () => { notifications++; });
  t.mock.method(globalThis, "fetch", async () => new Response("Current password is incorrect", { status: 401 }));
  await assert.rejects(auth.setPassword("new-password", "wrong-password"), /Current password is incorrect/);
  assert.equal(notifications, 0);
});
