const SECOND_WIDTH_REM = 4;

const timingDemo = document.querySelector(".timing-demo");
const timingUsernameField = document.querySelector("#timing-username");
const timingPasswordField = document.querySelector("#timing-password");
const timingButton = document.querySelector("#timing-go");
const timingAttemptsContainer = document.querySelector("#timing-attempts-container");
const timingRows = document.querySelector("#timing-attempts");
const timingServerCoverCheckbox = document.querySelector("#timing-hide-server");
const timingServerCover = document.querySelector("#timing-server-cover");

function TimingRow() {
  return E("tr", ["attempt"], [
    E("td"),
    E("td"),
    E("td", [], [
      Timeline(),
    ]),
  ]);
}

function Timeline() {
  function Tick(big) {
    return EAtts("div", ["bl"], { style: { width: `${SECOND_WIDTH_REM / 4}rem`, height: `${big ? 0.4 : 0.2}rem` } });
  }

  function Ticks() {
    return [ Tick(true), Tick(false), Tick(false), Tick(false) ];
  }

  return E("div", ["flex", "g2", "items-center"], [
    E("div", ["timer-time", "w2"], "0.0s"),
    E("div", ["flex", "flex-column"], [
      E("div", ["timer-playhead-spacer"]),
      E("div", ["relative"], [
        E("div", ["timer-playhead"]),
      ]),
      E("div", ["flex", "timer-bars"], [
        EAtts("div", ["timer-bar"], { style: { width: 0 } }),
      ]),
      E("div", ["bt", "flex", "items-start"], [
        ...Ticks(),
        ...Ticks(),
        ...Ticks(),
        EAtts("div", ["bl"], { style: { height: "0.4rem" } }),
      ]),
    ])
  ]);
}

function sleep(ms) {
  return new Promise((resolve, reject) => {
    setTimeout(resolve, ms);
  });
}

function frame(f) {
  return new Promise((resolve, reject) => {
    requestAnimationFrame(() => {
      try {
        f();
      } catch (e) {
        reject(e);
      }
      resolve();
    });
  });
}

class Timer {
  constructor(duration) {
    this.duration = duration;
    this.start = Date.now();
  }

  elapsed() {
    return Date.now() - this.start;
  }

  elapsedSec() {
    return this.elapsed() / 1000;
  }

  done() {
    return this.elapsed() >= this.duration;
  }
}

timingServerCoverCheckbox.addEventListener("change", e => {
  timingDemo.classList.toggle("hide-server", e.target.checked);
});
timingServerCoverCheckbox.checked = false;

timingButton.addEventListener("click", async () => {
  timingButton.setAttribute("disabled", "disabled");

  const inputUsername = timingUsernameField.value;
  const inputPassword = timingPasswordField.value;
  console.log("looking up", inputUsername);

  const stages = [];
  let response = timingMessageError;
  stages.push({ name: "send request", duration: 250, color: "bg-blue" });
  stages.push({ name: "database lookup", duration: 250, color: "bg-red" });
  stages.push({ name: "database lookup", duration: 500, color: "bg-red" });
  stages.push({ name: "database lookup", duration: 250, color: "bg-red" });
  const user = timingDb.find(user => user.username === inputUsername);
  if (user) {
    console.log("found user", user);
    stages.push({ name: "password check", duration: 1000, color: "bg-green" });
    if (user.password === inputPassword) {
      response = timingMessageSuccess;
    }
  }
  stages.push({ name: "receive response", duration: 250, color: "bg-blue" });

  // TODO: Find an empty one or whatever
  let row = timingDemo.querySelector(".attempt.empty");
  if (!row) {
    row = TimingRow();
    timingRows.appendChild(row);
  }
  row.classList.remove("empty");
  row.children[0].innerText = inputUsername;
  timingAttemptsContainer.scrollTo(0, timingAttemptsContainer.scrollHeight);

  const playhead = row.querySelector(".timer-playhead");
  const bars = row.querySelector(".timer-bars");
  const display = row.querySelector(".timer-time");
  let currentStageStartTime = 0;
  let currentBar = row.querySelector(".timer-bar");
  let currentStage = null;

  function newStage(start, stage) {
    const bar = E("div", ["timer-bar"]);
    bar.classList.add(stage.color);
    bars.appendChild(bar);
    currentStageStartTime = start;
    currentBar = bar;
    currentStage = stage;
  }

  bars.innerHTML = "";
  newStage(0, stages[0]);

  const timer = new Timer(10000);
  while (!timer.done()) {
    await frame(() => {
      playhead.style.left = `${timer.elapsedSec() * SECOND_WIDTH_REM}rem`;
      currentBar.style.width = `${(timer.elapsed() - currentStageStartTime) / 1000 * SECOND_WIDTH_REM}rem`;
      display.innerText = `${timer.elapsedSec().toFixed(1)}s`;
    });

    if (timer.elapsed() >= currentStageStartTime + currentStage.duration) {
      const newStageIndex = stages.indexOf(currentStage) + 1;
      if (newStageIndex >= stages.length) {
        break;
      }
      newStage(timer.elapsed(), stages[newStageIndex]);
    }
  }

  row.children[1].innerText = response;
  timingButton.removeAttribute("disabled");
});

for (const attempt of timingDemo.querySelectorAll(".attempt")) {
  attempt.children[2].appendChild(Timeline());
}
