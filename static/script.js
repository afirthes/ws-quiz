function startEnglishWordsQuiz(msg) {
    console.log(msg)
}

function joinQuiz() {
    console.log("joinQuiz")
}

function leaveQuiz() {
    console.log("leaveQuiz")
}

function startQuiz() {
    console.log("startQuiz")
}

function cancelQuiz() {
    console.log("cancelQuiz")
}

function submitAnswer() {
    console.log("submitAnswer")
    startProgress()
}

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