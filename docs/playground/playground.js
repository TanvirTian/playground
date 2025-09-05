window.playground = function() {
  const API = "https://playground-r6wl.onrender.com";

  return {
    code: "",
    lines: [],
    session: "",
    lastExecutedLine: 0, 

    async init() {
      try {
        const res = await fetch(`${API}/session`);
        const data = await res.json();
        this.session = data.token;
      } catch (err) {
        console.error("Failed to get session token:", err);
      }
    },

    append(input, output, ok = true) {
      this.lines.push({ id: crypto.randomUUID(), input, output, ok });
      this.$nextTick(() => {
        const el = document.getElementById("output");
        el.scrollTop = el.scrollHeight;
      });
    },

    async run() {
      const allLines = this.code.split(/\n+/).map(s => s.trim()).filter(Boolean);
      const blocks = allLines.slice(this.lastExecutedLine); // only new lines

      for (const b of blocks) {
        try {
          const res = await fetch(`${API}/eval`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': 'Bearer ' + this.session
            },
            body: JSON.stringify({ line: b })
          });
          const data = await res.json();

          const newToken = res.headers.get("X-Session-Token");
          if (newToken) this.session = newToken;

          this.append(b, data.ok ? String(data.result) : data.error, data.ok);
        } catch (err) {
          this.append(b, 'Network error', false);
        }
      }

      this.lastExecutedLine = allLines.length; 
    },

    clearOutput() {
      this.lines = [];
      this.lastExecutedLine = 0; 
    }
  };
};
