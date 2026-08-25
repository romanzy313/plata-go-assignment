async function main() {
  var res;
  var data;

  res = await fetch("http://localhost:3000/health");
  if (!res.ok) {
    throw Error("server not started");
  }

  const idempotencyKey = crypto.randomUUID();
  res = await fetch("http://localhost:3000/update", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": idempotencyKey,
    },
    body: JSON.stringify({ pair: "USD/MXN" }),
  });
  data = await res.json();
  console.log("update result", data);
  if (!res.ok) {
    throw Error("failed to update");
  }

  res = await fetch(`http://localhost:3000/quote/${data.updateId}`);
  data = await res.json();
  console.log("get by id", data);
  if (!res.ok) {
    throw Error("failed get by id");
  }

  res = await fetch(
    `http://localhost:3000/quote/latest?pair=${encodeURIComponent("USD/MXN")}`,
  );
  data = await res.json();
  console.log("get latest", data);
  if (res.ok) {
    throw Error("not expected");
  }
}

await main();
