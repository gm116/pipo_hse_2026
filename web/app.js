const state = {
  token: localStorage.getItem("token") || "",
  user: null,
  tasks: [],
};

const els = {
  authPanel: document.getElementById("authPanel"),
  appPanel: document.getElementById("appPanel"),
  authStatus: document.getElementById("authStatus"),
  taskStatus: document.getElementById("taskStatus"),
  userInfo: document.getElementById("userInfo"),
  registerForm: document.getElementById("registerForm"),
  loginForm: document.getElementById("loginForm"),
  taskForm: document.getElementById("taskForm"),
  tasks: document.getElementById("tasks"),
  logoutBtn: document.getElementById("logoutBtn"),
  reloadBtn: document.getElementById("reloadBtn"),
};

init();

function init() {
  bindEvents();
  if (state.token) {
    restoreSession();
  } else {
    showAuth();
  }
}

function bindEvents() {
  els.registerForm.addEventListener("submit", onRegister);
  els.loginForm.addEventListener("submit", onLogin);
  els.taskForm.addEventListener("submit", onCreateTask);
  els.logoutBtn.addEventListener("click", onLogout);
  els.reloadBtn.addEventListener("click", () => loadTasks());
}

async function restoreSession() {
  setStatus(els.authStatus, "Проверяю сессию...");
  try {
    state.user = await api("/api/auth/v1/users/me");
    showApp();
    await loadTasks();
  } catch (error) {
    onLogout();
    setStatus(els.authStatus, error.message, true);
  }
}

async function onRegister(event) {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(els.registerForm));
  setStatus(els.authStatus, "Регистрирую пользователя...");
  try {
    const result = await api("/api/auth/v1/users/register", {
      method: "POST",
      body: JSON.stringify(data),
    }, false);
    saveSession(result);
    els.registerForm.reset();
    showApp();
    await loadTasks();
    setStatus(els.authStatus, "");
  } catch (error) {
    setStatus(els.authStatus, error.message, true);
  }
}

async function onLogin(event) {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(els.loginForm));
  setStatus(els.authStatus, "Вхожу...");
  try {
    const result = await api("/api/auth/v1/users/login", {
      method: "POST",
      body: JSON.stringify(data),
    }, false);
    saveSession(result);
    els.loginForm.reset();
    showApp();
    await loadTasks();
    setStatus(els.authStatus, "");
  } catch (error) {
    setStatus(els.authStatus, error.message, true);
  }
}

function onLogout() {
  state.token = "";
  state.user = null;
  state.tasks = [];
  localStorage.removeItem("token");
  localStorage.removeItem("user");
  renderTasks();
  showAuth();
}

async function onCreateTask(event) {
  event.preventDefault();
  const data = Object.fromEntries(new FormData(els.taskForm));
  setStatus(els.taskStatus, "Создаю задачу...");
  try {
    await api("/api/tasks/v1/tasks", {
      method: "POST",
      body: JSON.stringify(data),
    });
    els.taskForm.reset();
    await loadTasks();
    setStatus(els.taskStatus, "Задача добавлена.");
  } catch (error) {
    setStatus(els.taskStatus, error.message, true);
  }
}

async function loadTasks() {
  setStatus(els.taskStatus, "Загружаю список...");
  try {
    const payload = await api("/api/tasks/v1/tasks");
    state.tasks = payload.items || [];
    renderTasks();
    setStatus(els.taskStatus, "");
  } catch (error) {
    setStatus(els.taskStatus, error.message, true);
  }
}

function renderTasks() {
  if (!state.tasks.length) {
    els.tasks.innerHTML = `<div class="card">Пока пусто. Добавь первую задачу.</div>`;
    return;
  }

  const cards = state.tasks.map((task) => {
    const created = new Date(task.createdAt).toLocaleString();
    return `
      <article class="task" data-id="${task.id}">
        <div class="task-head">
          <strong>${escapeHtml(task.title)}</strong>
          <span class="badge">${escapeHtml(task.status)}</span>
        </div>
        <p>${escapeHtml(task.description || "Без описания")}</p>
        <small>Создано: ${created}</small>
        <div class="task-actions">
          <button type="button" class="ghost" data-action="edit">Редактировать</button>
          <button type="button" class="ghost" data-action="delete">Удалить</button>
        </div>
      </article>
    `;
  });

  els.tasks.innerHTML = cards.join("");
  els.tasks.querySelectorAll("button").forEach((button) => {
    button.addEventListener("click", onTaskAction);
  });
}

async function onTaskAction(event) {
  const button = event.currentTarget;
  const article = button.closest(".task");
  const taskId = article?.dataset.id;
  if (!taskId) {
    return;
  }

  const action = button.dataset.action;
  const task = state.tasks.find((item) => String(item.id) === String(taskId));
  if (!task) {
    return;
  }

  if (action === "delete") {
    if (!window.confirm("Удалить задачу?")) {
      return;
    }
    try {
      await api(`/api/tasks/v1/tasks/${taskId}`, { method: "DELETE" });
      await loadTasks();
    } catch (error) {
      setStatus(els.taskStatus, error.message, true);
    }
    return;
  }

  if (action === "edit") {
    const title = window.prompt("Название", task.title);
    if (title === null) {
      return;
    }
    const description = window.prompt("Описание", task.description || "") ?? "";
    const status = (window.prompt("Статус (todo/in_progress/done)", task.status) || "").trim();

    try {
      await api(`/api/tasks/v1/tasks/${taskId}`, {
        method: "PUT",
        body: JSON.stringify({ title, description, status }),
      });
      await loadTasks();
    } catch (error) {
      setStatus(els.taskStatus, error.message, true);
    }
  }
}

function saveSession(result) {
  state.token = result.token;
  state.user = result.user;
  localStorage.setItem("token", result.token);
  localStorage.setItem("user", JSON.stringify(result.user));
}

function showAuth() {
  els.authPanel.classList.remove("hidden");
  els.appPanel.classList.add("hidden");
  setStatus(els.taskStatus, "");
}

function showApp() {
  els.authPanel.classList.add("hidden");
  els.appPanel.classList.remove("hidden");
  const user = state.user || JSON.parse(localStorage.getItem("user") || "null");
  if (user) {
    els.userInfo.textContent = `${user.name} (${user.email})`;
  }
}

async function api(url, options = {}, auth = true) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };
  if (auth && state.token) {
    headers.Authorization = `Bearer ${state.token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (response.status === 204) {
    return null;
  }

  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(payload.error || `HTTP ${response.status}`);
  }

  return payload;
}

function setStatus(node, text, isError = false) {
  node.textContent = text;
  node.classList.toggle("error", Boolean(isError));
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
