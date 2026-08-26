async function main() {
  let ok, data;
  [ok] = await health();
  if (!ok) {
    throw Error("server not started");
  }

  const idempotencyKey = crypto.randomUUID();
  const pair = "EUR/MXN";

  [ok, data] = await newUpdate(idempotencyKey, pair);
  console.log("new update", data);
  if (!ok) {
    throw Error("failed to update");
  }

  const updateId = data.updateId;

  [ok, data] = await getLatest(pair);
  console.log("latest", data);

  for (var i = 1; i <= 4; i++) {
    [ok, data] = await getById(updateId);
    console.log(`attempt ${i}: get by id`, data);
    if (!ok) {
      throw Error("failed get by id");
    }

    if (data.status == "completed") {
      console.log("e2e success!");
      return;
    }

    await sleep(5000);
  }

  throw new Error("e2e failed");
}

async function health() {
  const res = await fetch("http://localhost:3000/health");
  return [res.ok];
}

async function newUpdate(idempotencyKey, pair) {
  const res = await fetch(
    `http://localhost:3000/update?pair=${encodeURIComponent(pair)}`,
    {
      method: "POST",
      headers: {
        "Idempotency-Key": idempotencyKey,
      },
    },
  );
  const data = await res.json();
  return [res.ok, data];
}

async function getById(updateId) {
  const res = await fetch(`http://localhost:3000/quote/${updateId}`);
  const data = await res.json();
  return [res.ok, data];
}

async function getLatest(pair) {
  const res = await fetch(
    `http://localhost:3000/quote/latest?pair=${encodeURIComponent(pair)}`,
  );
  const data = await res.json();
  return [res.ok, data];
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// await main();
await Promise.all(Array.from({ length: 5 }).map(() => main()));
