

window.quizzesData = [
    {
        title: "Викторина на знание английских слов",
        questions: [
            {
                "question": "Translate to English: 'Como estas?'",
                "answers": [
                    "How are you?",
                    "Where are you?",
                    "Who are you?"
                ],
                "correct_answer": 1,
                "cost": 1000
            },
            {
                "question": "What is the English word for 'manzana' (in Spanish)?",
                "answers": [
                    "Banana",
                    "Apple",
                    "Grape"
                ],
                "correct_answer": 2,
                "cost": 1000
            },
            {
                "question": "Choose the correct translation: 'Je suis fatigué'",
                "answers": [
                    "I am hungry",
                    "I am tired",
                    "I am happy"
                ],
                "correct_answer": 2,
                "cost": 1000
            }
        ]
    }
]

let socket = null;

// Глобальный массив подписчиков на сообщения
const websocketListeners = [];

function registerWebSocketListener(callback) {
    websocketListeners.push(callback);
}

// document.addEventListener("DOMContentLoaded", function () {
//
//     // initializing
//     let userNameInput = document.getElementById("userNameInput")
//     let wsConnectButton = document.getElementById("ws-connect")
//     wsConnectButton.addEventListener("click", function () {
//         if (userNameInput.value.trim() === "") {
//             alert("Input is empty");
//         } else {
//             const userId = generate();
//             const userName = userNameInput.value;
//
//             connectWebSockets(userId, userName)
//         }
//     });
//
// })


function startProgress() {
    const progressBar = document.getElementById('progressBar');
    let width = 0;
    const interval = setInterval(() => {
        if (width >= 100) {
            clearInterval(interval);
        } else {
            width++;
            progressBar.style.width = width + '%';
        }
    }, 50); // обновлять каждую 50 мс, время заполнения ~5 секунд
}


function connectWebSockets(userId, userName) {
    socket = new ReconnectingWebSocket(`ws://localhost:8080/ws?user-id=${encodeURIComponent(userId)}&user-name=${encodeURIComponent(userName)}`, null, {
        debug: true,
        reconnectInterval: 3000
    });

    const offline = `<span class="bg-red-100 text-green-800 text-sm font-semibold px-3 py-1 rounded">Offline</span>`
    const online = `<span class="bg-green-100 text-green-800 text-sm font-semibold px-3 py-1 rounded">Connected</span>`
    let statusDiv = document.getElementById("status")

    socket.onopen = () => {
        console.log("Successfully connected")
        statusDiv.innerHTML = online
    }

    socket.onclose = () => {
        console.log("Connection closed")
        statusDiv.innerHTML = offline
    }

    socket.onerror = () => {
        console.log("There was an error.")
        // TODO: make consideration
        statusDiv.innerHTML = offline
    }

    socket.onmessage = msg => {
        console.log(msg.data)
        let data = JSON.parse(msg.data);
        console.log("Action:", data.action);
        websocketListeners.forEach(fn => fn(data));
    }

}

function sendMessage() {
    if (!socket) {
        errorMessage("no connection")
        return false
    }
    if ((userField.value.trim() === "") || (messageField.value.trim() === "")) {
        errorMessage("fill out username and message")
        return false
    }
    let jsonData = {}
    jsonData["action"] = "broadcast";
    jsonData["username"] = userField.value;
    jsonData["message"] = messageField.value;
    console.log("Sending message")
    socket.send(JSON.stringify(jsonData));
    messageField.value = "";
}

function errorMessage(msg) {
    notie.alert({
        type: 'error',
        text: msg,
    })
}
