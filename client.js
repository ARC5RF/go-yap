let YAP = { handlers: {} };

let socket_outcome = {};
let socket_open = new Promise((res, rej) => {
  socket_outcome.res = res;
  socket_outcome.rej = rej;
});

console.log("Attempting Connection...");
let proto = window.location.protocol == "https:" ? "wss://" : "ws://";
YAP.socket = new WebSocket(proto + window.location.host + "/yap/ws");
YAP.socket.onopen = function () {
  console.log("YAP Successfully Connected");
  socket_outcome.res(true);
};
YAP.socket.onclose = console.log.bind("YAP Socket Closed Connection: ");
YAP.socket.onerror = function (err) {
  console.log("YAP Socket Error: ", err);
  socket_outcome.res(false);
};
YAP.socket.onmessage = function (message) {
  let parsed = JSON.parse(message.data);

  let handler = YAP.handlers[parsed.Receiver];
  if (handler) handler(...parsed.Tokens);
};

YAP.subscribe = async function (reciever, handler, keep) {
  if (!(await socket_open)) return;

  let d = { Reciever: reciever, Keep: keep };
  let m = JSON.stringify({ Receiver: "yap.subscribe", Data: d });
  console.log(YAP.socket.send(m));
  YAP.handlers[reciever] = handler;
};

window.YAP = YAP;
