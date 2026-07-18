(() => {
  const form = document.getElementById("login-form");
  const emailInput = document.getElementById("email");
  const passwordInput = document.getElementById("password");
  const formError = document.getElementById("form-error");

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    formError.classList.add("hidden");

    try {
      const res = await fetch("/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: emailInput.value, password: passwordInput.value }),
      });

      if (!res.ok) {
        formError.textContent = "Invalid email or password.";
        formError.classList.remove("hidden");
        return;
      }

      window.location.href = "/dashboard/";
    } catch (err) {
      formError.textContent = `Failed to sign in: ${err.message}`;
      formError.classList.remove("hidden");
    }
  });
})();
