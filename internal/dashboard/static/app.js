(() => {
  const state = {
    project: null,
    flags: [],
  };

  const $ = (id) => document.getElementById(id);

  const projectInput = $("project-input");
  const projectList = $("project-list");
  const loadBtn = $("load-btn");
  const emptyState = $("empty-state");
  const board = $("board");
  const boardProject = $("board-project");
  const flagsTbody = $("flags-tbody");
  const flagsEmpty = $("flags-empty");
  const newFlagBtn = $("new-flag-btn");

  const flagModal = $("flag-modal");
  const modalTitle = $("modal-title");
  const flagForm = $("flag-form");
  const flagNameInput = $("flag-name");
  const flagActiveInput = $("flag-active");
  const flagPercentInput = $("flag-percent");
  const flagSessionsInput = $("flag-sessions");
  const formError = $("form-error");
  const modalCancel = $("modal-cancel");

  const settingsBtn = $("settings-btn");
  const tokenModal = $("token-modal");
  const tokenInput = $("service-token-input");
  const tokenSave = $("token-save");
  const tokenCancel = $("token-cancel");

  const themeToggleBtn = $("theme-toggle-btn");

  const TOKEN_KEY = "ff_dashboard_service_token";
  const LAST_PROJECT_KEY = "ff_dashboard_last_project";
  const THEME_KEY = "ff_dashboard_theme";
  let editingFlagName = null;

  function systemPrefersDark() {
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  }

  function currentEffectiveTheme() {
    const stored = localStorage.getItem(THEME_KEY);
    if (stored === "light" || stored === "dark") return stored;
    return systemPrefersDark() ? "dark" : "light";
  }

  function updateThemeToggleIcon() {
    const effective = currentEffectiveTheme();
    themeToggleBtn.textContent = effective === "dark" ? "🌙" : "☀️";
    themeToggleBtn.title = effective === "dark" ? "Switch to light theme" : "Switch to dark theme";
  }

  function applyTheme(theme) {
    if (theme === "light" || theme === "dark") {
      document.documentElement.setAttribute("data-theme", theme);
    } else {
      document.documentElement.removeAttribute("data-theme");
    }
    updateThemeToggleIcon();
  }

  themeToggleBtn.addEventListener("click", () => {
    const next = currentEffectiveTheme() === "dark" ? "light" : "dark";
    localStorage.setItem(THEME_KEY, next);
    applyTheme(next);
  });

  applyTheme(localStorage.getItem(THEME_KEY));

  function getServiceToken() {
    return localStorage.getItem(TOKEN_KEY) || "";
  }

  function authHeaders() {
    const token = getServiceToken();
    return token ? { Authorization: token } : {};
  }

  async function fetchJSON(url, options = {}) {
    const res = await fetch(url, options);
    if (!res.ok) {
      const text = await res.text().catch(() => "");
      throw new Error(text || `Request failed with status ${res.status}`);
    }
    if (res.status === 204) return null;
    const contentType = res.headers.get("content-type") || "";
    if (contentType.includes("application/json")) return res.json();
    return null;
  }

  async function loadProjects() {
    try {
      const projects = await fetchJSON("/featureflag/projects");
      projectList.innerHTML = "";
      (projects || []).forEach((p) => {
        const opt = document.createElement("option");
        opt.value = p;
        projectList.appendChild(opt);
      });
    } catch (e) {
      // best effort: project autocomplete is optional
      console.warn("failed to load projects", e);
    }
  }

  // Entity.Strategies (as returned by GET /featureflag/{project}/all) serializes
  // session_id as an object map (session -> bool), not an array.
  function sessionKeys(strategy) {
    return Object.keys(strategy.session_id || {});
  }

  function strategyDescription(flag) {
    const strategy = flag.strategy || {};
    const sessions = sessionKeys(strategy);
    if (sessions.length > 0) {
      return `Sessions (${sessions.length})`;
    }
    if (strategy.percent && strategy.percent > 0) {
      return `${strategy.percent}% rollout`;
    }
    return "No strategy";
  }

  function renderFlags() {
    flagsTbody.innerHTML = "";

    if (state.flags.length === 0) {
      flagsEmpty.classList.remove("hidden");
      return;
    }
    flagsEmpty.classList.add("hidden");

    state.flags
      .slice()
      .sort((a, b) => a.flag_name.localeCompare(b.flag_name))
      .forEach((flag) => {
        const tr = document.createElement("tr");

        const nameTd = document.createElement("td");
        nameTd.className = "flag-name";
        nameTd.textContent = flag.flag_name;
        tr.appendChild(nameTd);

        const statusTd = document.createElement("td");
        const pill = document.createElement("span");
        pill.className = `status-pill ${flag.active ? "active" : "inactive"}`;
        pill.textContent = flag.active ? "Active" : "Inactive";
        statusTd.appendChild(pill);
        tr.appendChild(statusTd);

        const strategyTd = document.createElement("td");
        strategyTd.className = "strategy-desc";
        strategyTd.textContent = strategyDescription(flag);
        tr.appendChild(strategyTd);

        const actionsTd = document.createElement("td");
        actionsTd.className = "row-actions";

        const toggleBtn = document.createElement("button");
        toggleBtn.textContent = flag.active ? "Deactivate" : "Activate";
        toggleBtn.addEventListener("click", () => toggleFlag(flag));
        actionsTd.appendChild(toggleBtn);

        const editBtn = document.createElement("button");
        editBtn.textContent = "Edit";
        editBtn.addEventListener("click", () => openEditModal(flag));
        actionsTd.appendChild(editBtn);

        const deleteBtn = document.createElement("button");
        deleteBtn.textContent = "Delete";
        deleteBtn.className = "danger";
        deleteBtn.addEventListener("click", () => deleteFlag(flag));
        actionsTd.appendChild(deleteBtn);

        tr.appendChild(actionsTd);
        flagsTbody.appendChild(tr);
      });
  }

  async function loadFlags(project) {
    const flags = await fetchJSON(`/featureflag/${encodeURIComponent(project)}/all`);
    state.project = project;
    state.flags = flags || [];
    boardProject.textContent = project;
    emptyState.classList.add("hidden");
    board.classList.remove("hidden");
    renderFlags();
    localStorage.setItem(LAST_PROJECT_KEY, project);
  }

  // Converts the domain-shaped strategy from GET /all (session_id as a map)
  // into the wire DTO shape the PATCH endpoint expects (session_id as an array).
  function toStrategyDtoPayload(strategy) {
    strategy = strategy || {};
    return {
      percent: strategy.percent || 0,
      session_id: sessionKeys(strategy),
    };
  }

  async function toggleFlag(flag) {
    try {
      await saveFlag({
        flag_name: flag.flag_name,
        active: !flag.active,
        strategy: toStrategyDtoPayload(flag.strategy),
      });
      await loadFlags(state.project);
    } catch (e) {
      alert(`Failed to update flag: ${e.message}`);
    }
  }

  async function deleteFlag(flag) {
    if (!confirm(`Delete flag "${flag.flag_name}"?`)) return;

    const token = getServiceToken();
    if (!token) {
      alert("Set the service token (gear icon) to delete flags.");
      openTokenModal();
      return;
    }

    try {
      await fetchJSON(`/featureflag/${encodeURIComponent(state.project)}/${encodeURIComponent(flag.flag_name)}`, {
        method: "DELETE",
        headers: authHeaders(),
      });
      await loadFlags(state.project);
    } catch (e) {
      alert(`Failed to delete flag: ${e.message}`);
    }
  }

  function strategyPayload() {
    const percent = parseFloat(flagPercentInput.value);
    const sessions = flagSessionsInput.value
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

    return {
      percent: Number.isFinite(percent) ? percent : 0,
      session_id: sessions,
    };
  }

  async function saveFlag(payload) {
    await fetchJSON(`/featureflag/${encodeURIComponent(state.project)}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
  }

  function openNewModal() {
    editingFlagName = null;
    modalTitle.textContent = "New feature flag";
    flagForm.reset();
    flagNameInput.disabled = false;
    formError.classList.add("hidden");
    flagModal.classList.remove("hidden");
    flagNameInput.focus();
  }

  function openEditModal(flag) {
    editingFlagName = flag.flag_name;
    modalTitle.textContent = `Edit "${flag.flag_name}"`;
    flagNameInput.value = flag.flag_name;
    flagNameInput.disabled = true;
    flagActiveInput.checked = !!flag.active;
    const strategy = flag.strategy || {};
    flagPercentInput.value = strategy.percent || "";
    flagSessionsInput.value = sessionKeys(strategy).join(", ");
    formError.classList.add("hidden");
    flagModal.classList.remove("hidden");
  }

  function closeFlagModal() {
    flagModal.classList.add("hidden");
  }

  function openTokenModal() {
    tokenInput.value = getServiceToken();
    tokenModal.classList.remove("hidden");
  }

  function closeTokenModal() {
    tokenModal.classList.add("hidden");
  }

  loadBtn.addEventListener("click", async () => {
    const project = projectInput.value.trim();
    if (!project) return;
    try {
      await loadFlags(project);
      await loadProjects();
    } catch (e) {
      alert(`Failed to load project: ${e.message}`);
    }
  });

  projectInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      loadBtn.click();
    }
  });

  newFlagBtn.addEventListener("click", openNewModal);
  modalCancel.addEventListener("click", closeFlagModal);
  flagModal.addEventListener("click", (e) => {
    if (e.target === flagModal) closeFlagModal();
  });

  flagForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    formError.classList.add("hidden");

    const flagName = editingFlagName || flagNameInput.value.trim();
    if (!flagName) return;

    const payload = {
      flag_name: flagName,
      active: flagActiveInput.checked,
      strategy: strategyPayload(),
    };

    try {
      await saveFlag(payload);
      closeFlagModal();
      await loadFlags(state.project);
    } catch (err) {
      formError.textContent = err.message;
      formError.classList.remove("hidden");
    }
  });

  settingsBtn.addEventListener("click", openTokenModal);
  tokenCancel.addEventListener("click", closeTokenModal);
  tokenModal.addEventListener("click", (e) => {
    if (e.target === tokenModal) closeTokenModal();
  });
  tokenSave.addEventListener("click", () => {
    localStorage.setItem(TOKEN_KEY, tokenInput.value.trim());
    closeTokenModal();
  });

  async function init() {
    await loadProjects();

    const lastProject = localStorage.getItem(LAST_PROJECT_KEY);
    if (!lastProject) return;

    projectInput.value = lastProject;
    try {
      await loadFlags(lastProject);
    } catch (e) {
      // best effort: fall back to the empty state if the remembered project is gone
      console.warn("failed to auto-load last project", e);
    }
  }

  init();
})();
