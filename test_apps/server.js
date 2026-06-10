import net from "net";

const client = net.createConnection({ host: "127.0.0.1", port: 6379 }, () => {
	client.write("SET name ted\r\n");
	client.write("GET name");
});

client.on("data", (data) => {
	console.log(data.toString());
});

client.on("end", () => {
	console.log("disconnected");
});
