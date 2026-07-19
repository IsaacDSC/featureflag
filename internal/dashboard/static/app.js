(() => {
  const state = {
    project: null,
    flags: [],
    users: [],
    availableProjects: [],
  };

  const $ = (id) => document.getElementById(id);

  const navFlagsBtn = $("nav-flags-btn");
  const navUsersBtn = $("nav-users-btn");
  const projectPicker = $("project-picker");
  const flagsView = $("flags-view");
  const usersView = $("users-view");

  const projectInput = $("project-input");
  const projectList = $("project-list");
  const loadBtn = $("load-btn");
  const emptyState = $("empty-state");
  const board = $("board");
  const boardProject = $("board-project");
  const flagsTbody = $("flags-tbody");
  const flagsEmpty = $("flags-empty");
  const newFlagBtn = $("new-flag-btn");
  const logoutBtn = $("logout-btn");

  const usersTbody = $("users-tbody");
  const usersEmpty = $("users-empty");
  const newUserBtn = $("new-user-btn");
  const userModal = $("user-modal");
  const userModalTitle = $("user-modal-title");
  const userForm = $("user-form");
  const userEmailInput = $("user-email");
  const userPasswordRow = $("user-password-row");
  const userPasswordInput = $("user-password");
  const userRoleSelect = $("user-role");
  const userProjectsField = $("user-projects-field");
  const userProjectsList = $("user-projects-list");
  const userFormError = $("user-form-error");
  const userModalCancel = $("user-modal-cancel");
  const userFormSubmit = $("user-form-submit");

  const flagModal = $("flag-modal");
  const modalTitle = $("modal-title");
  const flagForm = $("flag-form");
  const flagNameInput = $("flag-name");
  const flagActiveInput = $("flag-active");
  const flagPercentInput = $("flag-percent");
  const flagSessionsInput = $("flag-sessions");
  const formError = $("form-error");
  const modalCancel = $("modal-cancel");

  const themeToggleBtn = $("theme-toggle-btn");

  const forcePasswordModal = $("force-password-modal");
  const forcePasswordForm = $("force-password-form");
  const forceCurrentPasswordInput = $("force-current-password");
  const forceNewPasswordInput = $("force-new-password");
  const forceConfirmPasswordInput = $("force-confirm-password");
  const forcePasswordError = $("force-password-error");

  const LAST_PROJECT_KEY = "ff_dashboard_last_project";
  const THEME_KEY = "ff_dashboard_theme";
  let editingFlagName = null;
  let editingUserEmail = null;

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
      state.availableProjects = projects || [];
      projectList.innerHTML = "";
      state.availableProjects.forEach((p) => {
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

    try {
      await fetchJSON(`/featureflag/${encodeURIComponent(state.project)}/${encodeURIComponent(flag.flag_name)}`, {
        method: "DELETE",
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

  function switchView(view) {
    const showUsers = view === "users";
    usersView.classList.toggle("hidden", !showUsers);
    flagsView.classList.toggle("hidden", showUsers);
    projectPicker.classList.toggle("hidden", showUsers);
    navUsersBtn.classList.toggle("active", showUsers);
    navFlagsBtn.classList.toggle("active", !showUsers);

    if (showUsers) {
      loadUsers().catch((e) => console.warn("failed to load users", e));
    }
  }

  function accessDescription(user) {
    if (user.role === "admin") return "Admin (all projects)";
    const projects = user.projects || [];
    return projects.length > 0 ? `Member: ${projects.join(", ")}` : "Member (no projects yet)";
  }

  function renderUsers() {
    usersTbody.innerHTML = "";

    if (state.users.length === 0) {
      usersEmpty.classList.remove("hidden");
      return;
    }
    usersEmpty.classList.add("hidden");

    state.users
      .slice()
      .sort((a, b) => a.email.localeCompare(b.email))
      .forEach((user) => {
        const tr = document.createElement("tr");

        const emailTd = document.createElement("td");
        emailTd.className = "flag-name";
        emailTd.textContent = user.email;
        tr.appendChild(emailTd);

        const accessTd = document.createElement("td");
        const badge = document.createElement("span");
        badge.className = "access-badge";
        badge.textContent = accessDescription(user);
        accessTd.appendChild(badge);
        tr.appendChild(accessTd);

        const createdTd = document.createElement("td");
        createdTd.className = "strategy-desc";
        createdTd.textContent = user.created_at ? new Date(user.created_at).toLocaleString() : "";
        tr.appendChild(createdTd);

        const actionsTd = document.createElement("td");
        actionsTd.className = "row-actions";

        const editBtn = document.createElement("button");
        editBtn.textContent = "Edit";
        editBtn.addEventListener("click", () => openEditUserModal(user));
        actionsTd.appendChild(editBtn);

        if (user.must_change_password) {
          const pendingBadge = document.createElement("span");
          pendingBadge.className = "access-badge";
          pendingBadge.textContent = "Password change pending";
          actionsTd.appendChild(pendingBadge);
        } else {
          const forceResetBtn = document.createElement("button");
          forceResetBtn.textContent = "Force password reset";
          forceResetBtn.addEventListener("click", () => forcePasswordReset(user));
          actionsTd.appendChild(forceResetBtn);
        }

        const deleteBtn = document.createElement("button");
        deleteBtn.textContent = "Delete";
        deleteBtn.className = "danger";
        deleteBtn.addEventListener("click", () => deleteUser(user));
        actionsTd.appendChild(deleteBtn);

        tr.appendChild(actionsTd);
        usersTbody.appendChild(tr);
      });
  }

  async function loadUsers() {
    const users = await fetchJSON("/users");
    state.users = users || [];
    renderUsers();
  }

  async function forcePasswordReset(user) {
    if (!confirm(`Force "${user.email}" to change their password on next login?`)) return;

    try {
      await fetchJSON(`/users/${encodeURIComponent(user.email)}/require-password-change`, { method: "POST" });
      await loadUsers();
    } catch (e) {
      alert(`Failed to force password reset: ${e.message}`);
    }
  }

  async function deleteUser(user) {
    if (!confirm(`Delete user "${user.email}"?`)) return;

    try {
      await fetchJSON(`/users/${encodeURIComponent(user.email)}`, { method: "DELETE" });
      await loadUsers();
    } catch (e) {
      alert(`Failed to delete user: ${e.message}`);
    }
  }

  function updateProjectsFieldVisibility() {
    userProjectsField.classList.toggle("hidden", userRoleSelect.value === "admin");
  }

  function renderProjectCheckboxes(selectedProjects) {
    const selected = new Set(selectedProjects || []);
    userProjectsList.innerHTML = "";

    if (state.availableProjects.length === 0) {
      const empty = document.createElement("p");
      empty.className = "no-projects";
      empty.textContent = "No projects yet — create a feature flag for a project first.";
      userProjectsList.appendChild(empty);
      return;
    }

    state.availableProjects.forEach((project) => {
      const label = document.createElement("label");
      const checkbox = document.createElement("input");
      checkbox.type = "checkbox";
      checkbox.value = project;
      checkbox.checked = selected.has(project);
      label.appendChild(checkbox);
      label.appendChild(document.createTextNode(project));
      userProjectsList.appendChild(label);
    });
  }

  function selectedProjectCheckboxes() {
    return Array.from(userProjectsList.querySelectorAll("input[type=checkbox]:checked")).map((cb) => cb.value);
  }

  function openNewUserModal() {
    editingUserEmail = null;
    userModalTitle.textContent = "New user";
    userFormSubmit.textContent = "Create";
    userForm.reset();
    userEmailInput.disabled = false;
    userPasswordRow.classList.remove("hidden");
    userPasswordInput.required = true;
    userRoleSelect.value = "member";
    renderProjectCheckboxes([]);
    updateProjectsFieldVisibility();
    userFormError.classList.add("hidden");
    userModal.classList.remove("hidden");
    userEmailInput.focus();
  }

  function openEditUserModal(user) {
    editingUserEmail = user.email;
    userModalTitle.textContent = `Edit "${user.email}"`;
    userFormSubmit.textContent = "Save";
    userForm.reset();
    userEmailInput.value = user.email;
    userEmailInput.disabled = true;
    userPasswordRow.classList.add("hidden");
    userPasswordInput.required = false;
    userRoleSelect.value = user.role === "admin" ? "admin" : "member";
    renderProjectCheckboxes(user.projects);
    updateProjectsFieldVisibility();
    userFormError.classList.add("hidden");
    userModal.classList.remove("hidden");
  }

  function closeUserModal() {
    userModal.classList.add("hidden");
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

  logoutBtn.addEventListener("click", async () => {
    try {
      await fetchJSON("/auth/logout", { method: "POST" });
    } catch (e) {
      // best effort: even if the request fails, send the user back to login
      console.warn("logout request failed", e);
    }
    window.location.href = "/auth/login/";
  });

  navFlagsBtn.addEventListener("click", () => switchView("flags"));
  navUsersBtn.addEventListener("click", () => switchView("users"));

  newUserBtn.addEventListener("click", openNewUserModal);
  userModalCancel.addEventListener("click", closeUserModal);
  userModal.addEventListener("click", (e) => {
    if (e.target === userModal) closeUserModal();
  });
  userRoleSelect.addEventListener("change", updateProjectsFieldVisibility);

  userForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    userFormError.classList.add("hidden");

    const role = userRoleSelect.value;
    const projects = role === "admin" ? [] : selectedProjectCheckboxes();

    try {
      if (editingUserEmail) {
        await fetchJSON(`/users/${encodeURIComponent(editingUserEmail)}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ role, projects }),
        });
      } else {
        await fetchJSON("/users", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            email: userEmailInput.value.trim(),
            password: userPasswordInput.value,
            role,
            projects,
          }),
        });
      }
      closeUserModal();
      await loadUsers();
    } catch (err) {
      userFormError.textContent = err.message;
      userFormError.classList.remove("hidden");
    }
  });

  // applyCurrentUserUI is best effort only: the server is the real gate
  // (403 on /users for non-admins) — this just avoids showing a tab that
  // would 403.
  function applyCurrentUserUI(me) {
    if (me && me.role !== "admin") {
      navUsersBtn.classList.add("hidden");
    }
  }

  async function loadDashboard() {
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

  function showForcePasswordModal() {
    forcePasswordError.classList.add("hidden");
    forcePasswordForm.reset();
    forcePasswordModal.classList.remove("hidden");
    forceCurrentPasswordInput.focus();
  }

  function hideForcePasswordModal() {
    forcePasswordModal.classList.add("hidden");
  }

  // The dashboard itself stays behind the modal-backdrop (no cancel button,
  // no close-on-backdrop-click) until this succeeds — every other route is
  // also 403'd server-side (requirePasswordChanged) while must_change_password
  // is true, so this is UX, not the actual security gate.
  forcePasswordForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    forcePasswordError.classList.add("hidden");

    if (forceNewPasswordInput.value !== forceConfirmPasswordInput.value) {
      forcePasswordError.textContent = "New password and confirmation do not match.";
      forcePasswordError.classList.remove("hidden");
      return;
    }

    try {
      await fetchJSON("/me/password", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          current_password: forceCurrentPasswordInput.value,
          new_password: forceNewPasswordInput.value,
        }),
      });
      hideForcePasswordModal();

      let me = null;
      try {
        me = await fetchJSON("/me");
      } catch (err) {
        console.warn("failed to load current user", err);
      }
      applyCurrentUserUI(me);
      await loadDashboard();
    } catch (err) {
      forcePasswordError.textContent = err.message;
      forcePasswordError.classList.remove("hidden");
    }
  });

  async function init() {
    let me = null;
    try {
      me = await fetchJSON("/me");
    } catch (e) {
      console.warn("failed to load current user", e);
    }

    if (me && me.must_change_password) {
      showForcePasswordModal();
      return;
    }

    applyCurrentUserUI(me);
    await loadDashboard();
  }

  init();
})();
