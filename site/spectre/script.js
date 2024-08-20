const SECOND_WIDTH_REM = 4;

const timingDemo = document.querySelector(".timing-demo");
const timingUsernameField = document.querySelector("#timing-username");
const timingPasswordField = document.querySelector("#timing-password");
const timingButton = document.querySelector("#timing-go");
const timingAttemptsContainer = document.querySelector("#timing-attempts-container");
const timingRows = document.querySelector("#timing-attempts");
const timingServerCoverCheckbox = document.querySelector("#timing-hide-server");
const timingServerCover = document.querySelector("#timing-server-cover");

const cpuSecret = [1,3,3,7];
const cpuSecretIndex = document.querySelector("#cpu-secret-index");
const cpuLookup = document.querySelector("#cpu-lookup");
const cpuProbe = document.querySelector("#cpu-probe");
const cpuClearCache = document.querySelector("#cpu-clear-cache");
const cpuUnlockVault = document.querySelector("#cpu-unlock-vault");
const cpuVaultDigit0 = document.querySelector("#cpu-vault-0");
const cpuVaultDigit1 = document.querySelector("#cpu-vault-1");
const cpuVaultDigit2 = document.querySelector("#cpu-vault-2");
const cpuVaultDigit3 = document.querySelector("#cpu-vault-3");
const cpuVaultPadlockUnlocked = document.querySelector("#cpu-vault-padlock-unlocked");
const cpuVaultPadlockLocked = document.querySelector("#cpu-vault-padlock-locked"); // cpu-vault-padlock-locked
const cpuCoverCheckbox = document.querySelector("#cpu-hide");
const cpuCPUCover = document.querySelector("#cpu-cover");

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


cpuLookup.addEventListener('click', e => {
  e.preventDefault();
  let secretIndex = cpuSecretIndex.value;
  if (secretIndex === "") {
      return;
  }
  secretIndex = +secretIndex;
  if (secretIndex < 0 || secretIndex > 6) {
      return;
  }
  const probeArrayIndex = cpuSecret[secretIndex];
  if (probeArrayIndex) {
      const cacheCell = document.querySelector("#cpu-cache-" + probeArrayIndex);
      cacheCell.innerText = "probeArray[" + probeArrayIndex + "]";
      const ramCell = document.querySelector("#cpu-ram-" + probeArrayIndex);
      ramCell.classList.add("bg-orange");
      setTimeout(()=> {
          ramCell.classList.remove("bg-orange");
      }, 300);
  }
});
cpuClearCache.addEventListener('click', e => {
  e.preventDefault();
  for (let i = 0; i <= 9; i += 1) {
      const cacheCell = document.querySelector("#cpu-cache-" + i);
      cacheCell.innerText = "";
  }
});
cpuProbe.addEventListener('click', async (e) => {
  e.preventDefault();
  cpuProbe.setAttribute("disabled", "disabled");
  const inCPUCache = [];
  for (let i = 0; i <= 9; i += 1) {
      if(document.querySelector("#cpu-cache-" + i).innerText.trim() !== "") {
          inCPUCache.push(i);
      }
  }
  let cpuProbeRows = document.querySelectorAll(".cpu-probe");
  for (let i = 0; i <= 9; i += 1) {
      const stages = [];
      const speedUp = 4;
      stages.push({name: "cpu_cache_lookup", duration: 500/speedUp, color: "bg-green"});
      if (!inCPUCache.includes(i)) {
          stages.push({name: "ram_lookup", duration: 1500/speedUp, color: "bg-red" });
      }
      const row = cpuProbeRows[i];
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
              playhead.style.left = `${timer.elapsedSec() * speedUp * SECOND_WIDTH_REM}rem`;
              currentBar.style.width = `${(timer.elapsed() - currentStageStartTime) * speedUp / 1000 * SECOND_WIDTH_REM}rem`;
              display.innerText = `${(timer.elapsedSec() * speedUp).toFixed(1)}s`;
          });
          if (timer.elapsed() >= currentStageStartTime + currentStage.duration) {
              const newStageIndex = stages.indexOf(currentStage) + 1;
              if (newStageIndex >= stages.length) {
                  break;
              }
              newStage(timer.elapsed(), stages[newStageIndex]);
          }
      }
  }
  cpuProbe.removeAttribute("disabled");
});
cpuUnlockVault.addEventListener('click', e => {
  e.preventDefault();
  let digit0 = cpuVaultDigit0.value;
  let digit1 = cpuVaultDigit1.value;
  let digit2 = cpuVaultDigit2.value;
  let digit3 = cpuVaultDigit3.value;
  if (digit0 === "" || digit1 === "" || digit2 === "" || digit3 === "") {
      return;
  }
  digit0 = +digit0;
  digit1 = +digit1;
  digit2 = +digit2;
  digit3 = +digit3;
  if (digit0 === cpuSecret[0] && digit1 === cpuSecret[1] && digit2 === cpuSecret[2] && digit3 === cpuSecret[3]) {
    console.log(document.getElementById("cpu-vault-padlock-locked"));
      cpuVaultPadlockLocked.classList.add("dn-l"); // Hide
      cpuVaultPadlockUnlocked.classList.remove("dn-l"); // Show
  }
});
cpuCoverCheckbox.addEventListener('change', e => {
  cpuCPUCover.style.opacity = e.target.checked? 1 : 0;
});

for (const cpuProbe of document.querySelectorAll(".cpu-probe")) {
  cpuProbe.children[1].appendChild(Timeline());
}

