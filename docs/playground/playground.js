window.playground = function() {
  const API = "https://playground-r6wl.onrender.com";

  return {
    code: "",
    lines: [],
    session: "",
    readonly: false, 
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
      if (this.readonly) return; 

      const allLines = this.code.split("\n"); 
      const blocks = allLines.slice(this.lastExecutedLine); 

      for (const b of blocks) {
        const trimmed = b.trim();
        if (!trimmed) {
          this.lastExecutedLine++;
          continue;
        }

        try {
          const res = await fetch(`${API}/eval`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': 'Bearer ' + this.session
            },
            body: JSON.stringify({ line: trimmed })
          });
          const data = await res.json();

          const newToken = res.headers.get("X-Session-Token");
          if (newToken) this.session = newToken;

          this.append(trimmed, data.ok ? String(data.result) : data.error, data.ok);
        } catch (err) {
          this.append(trimmed, 'Network error', false);
        }

        this.lastExecutedLine++; 
      }
    },

    clearOutput() {
      this.lines = [];
      this.lastExecutedLine = 0; 
    }
  };
};
