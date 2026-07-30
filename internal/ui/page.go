package ui

const pageHTML = `<!doctype html>
<html lang="tr">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Chat2API Session Connector</title>
  <style nonce="{{.Nonce}}">
    :root {
      color-scheme: dark;
      --bg: #07100f;
      --panel: #0d1816;
      --panel-2: #111f1c;
      --line: rgba(190, 231, 218, .13);
      --text: #edf7f3;
      --muted: #91aaa2;
      --accent: #44d7aa;
      --accent-2: #b8f2df;
      --danger: #ff8797;
      font-family: "Segoe UI", ui-sans-serif, system-ui, sans-serif;
    }
    * { box-sizing: border-box; }
    body {
      min-height: 100vh;
      margin: 0;
      display: grid;
      place-items: center;
      padding: 32px 18px;
      color: var(--text);
      background:
        radial-gradient(circle at 20% 0%, rgba(68, 215, 170, .12), transparent 34rem),
        radial-gradient(circle at 100% 100%, rgba(44, 112, 96, .15), transparent 32rem),
        var(--bg);
    }
    main { width: min(720px, 100%); }
    .brand { display: flex; align-items: center; gap: 13px; margin: 0 0 18px; }
    .mark {
      width: 42px; height: 42px; display: grid; place-items: center;
      border: 1px solid rgba(68, 215, 170, .3); border-radius: 13px;
      color: #04110d; background: linear-gradient(145deg, var(--accent-2), var(--accent));
      font-weight: 900; box-shadow: 0 12px 34px rgba(68, 215, 170, .12);
    }
    .brand strong { display: block; font-size: 15px; letter-spacing: -.01em; }
    .brand span { display: block; margin-top: 3px; color: var(--muted); font-size: 12px; }
    .card {
      overflow: hidden;
      border: 1px solid var(--line);
      border-radius: 24px;
      background: linear-gradient(155deg, rgba(17, 31, 28, .97), rgba(10, 19, 17, .98));
      box-shadow: 0 30px 90px rgba(0, 0, 0, .35);
    }
    .hero { padding: 30px 30px 25px; border-bottom: 1px solid var(--line); }
    .eyebrow { color: var(--accent); font-size: 11px; font-weight: 800; letter-spacing: .13em; text-transform: uppercase; }
    h1 { margin: 10px 0 9px; font-size: clamp(25px, 4vw, 36px); line-height: 1.08; letter-spacing: -.04em; }
    .hero p { max-width: 590px; margin: 0; color: var(--muted); font-size: 14px; line-height: 1.7; }
    .content { padding: 28px 30px 31px; }
    .steps { display: flex; align-items: center; margin-bottom: 24px; }
    .step { min-width: 0; display: flex; align-items: center; gap: 8px; color: #668078; font-size: 11px; font-weight: 700; }
    .step i {
      width: 25px; height: 25px; display: grid; place-items: center; flex: 0 0 auto;
      border: 1px solid var(--line); border-radius: 50%; font-style: normal;
    }
    .step.active { color: var(--text); }
    .step.active i, .step.complete i { color: #04110d; border-color: var(--accent); background: var(--accent); }
    .steps b { height: 1px; flex: 1; margin: 0 10px; background: var(--line); }
    label { display: block; margin-bottom: 8px; font-size: 12px; font-weight: 750; }
    textarea {
      width: 100%; min-height: 104px; resize: vertical;
      border: 1px solid var(--line); border-radius: 14px; outline: none;
      padding: 14px; color: var(--text); background: rgba(2, 9, 7, .55);
      font: 12px/1.55 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    }
    textarea:focus { border-color: rgba(68, 215, 170, .55); box-shadow: 0 0 0 4px rgba(68, 215, 170, .08); }
    .field-actions, .actions { display: flex; gap: 10px; margin-top: 13px; }
    button {
      min-height: 42px; display: inline-flex; align-items: center; justify-content: center;
      border: 0; border-radius: 12px; padding: 0 17px; cursor: pointer;
      color: #04110d; background: var(--accent); font: inherit; font-size: 12px; font-weight: 800;
    }
    button:hover { filter: brightness(1.05); }
    button:disabled { cursor: not-allowed; filter: grayscale(.5); opacity: .45; }
    button.secondary { color: var(--text); border: 1px solid var(--line); background: transparent; }
    button.ghost { color: var(--muted); background: transparent; }
    .preview, .status {
      margin-top: 18px; padding: 17px;
      border: 1px solid var(--line); border-radius: 15px; background: rgba(4, 12, 10, .45);
    }
    .preview small, .status small { color: var(--muted); }
    .host { margin: 6px 0 4px; overflow-wrap: anywhere; font-size: 16px; font-weight: 800; }
    .status { display: flex; align-items: flex-start; gap: 12px; }
    .status-dot { width: 9px; height: 9px; margin-top: 5px; flex: 0 0 auto; border-radius: 50%; background: var(--accent); box-shadow: 0 0 0 5px rgba(68, 215, 170, .08); }
    .status[data-phase="error"] .status-dot { background: var(--danger); box-shadow: 0 0 0 5px rgba(255, 135, 151, .08); }
    .status[data-phase="connecting"] .status-dot { animation: pulse 1.2s infinite; }
    .status strong { display: block; margin-bottom: 4px; font-size: 13px; }
    .status p { margin: 0; color: var(--muted); font-size: 12px; line-height: 1.55; }
    .privacy { margin: 22px 0 0; padding-top: 18px; border-top: 1px solid var(--line); color: #708981; font-size: 11px; line-height: 1.65; }
    [hidden] { display: none !important; }
    @keyframes pulse { 50% { opacity: .35; transform: scale(.72); } }
    @media (max-width: 620px) {
      body { padding: 14px; place-items: start center; }
      .card { border-radius: 19px; }
      .hero, .content { padding: 23px 20px; }
      .step span { display: none; }
      .field-actions, .actions { flex-direction: column; }
      button { width: 100%; }
    }
  </style>
</head>
<body>
  <main>
    <div class="brand"><div class="mark">C2</div><div><strong>Chat2API</strong><span>DeepSeek Session Connector</span></div></div>
    <section class="card">
      <div class="hero">
        <div class="eyebrow">Yerel ve güvenli bağlantı</div>
        <h1>DeepSeek hesabınızı bağlayın</h1>
        <p>Giriş doğrudan DeepSeek üzerinde tamamlanır. Connector parola veya doğrulama kodunuzu görmez; yalnız doğrulanmış oturum bilgisini seçtiğiniz Chat2API gateway’ine iletir.</p>
      </div>
      <div class="content">
        <div class="steps">
          <div class="step active" id="step-code"><i>1</i><span>Kod</span></div><b></b>
          <div class="step" id="step-confirm"><i>2</i><span>Onay</span></div><b></b>
          <div class="step" id="step-login"><i>3</i><span>Giriş</span></div><b></b>
          <div class="step" id="step-done"><i>4</i><span>Tamam</span></div>
        </div>

        <section id="code-view">
          <label for="code">Tek kullanımlık bağlantı kodu</label>
          <textarea id="code" autocomplete="off" autocapitalize="off" spellcheck="false" placeholder="c2a-ds-native-v1..."></textarea>
          <div class="field-actions">
            <button id="inspect" type="button">Kodu doğrula</button>
            <button id="paste" type="button" class="secondary">Panodan yapıştır</button>
          </div>
        </section>

        <section id="confirm-view" hidden>
          <div class="preview">
            <small>Oturumun gönderileceği gateway</small>
            <div class="host" id="gateway-host"></div>
            <small id="expiry"></small>
          </div>
          <div class="actions">
            <button id="connect" type="button">DeepSeek’i aç ve bağla</button>
            <button id="back" type="button" class="secondary">Kodu değiştir</button>
          </div>
        </section>

        <section id="status-view" hidden>
          <div class="status" id="status-card">
            <span class="status-dot"></span>
            <div><strong id="status-title"></strong><p id="status-message"></p></div>
          </div>
          <div class="actions" id="status-actions">
            <button id="retry" type="button" class="secondary" hidden>Tekrar dene</button>
            <button id="close" type="button" class="ghost" hidden>Connector’ı kapat</button>
          </div>
        </section>

        <p class="privacy">Kişisel tarayıcı profiliniz kullanılmaz. Geçici profil işlem sonunda silinir; token diske veya uygulama loglarına yazılmaz.</p>
      </div>
    </section>
  </main>
  <script nonce="{{.Nonce}}">
    const base = {{printf "%q" .BasePath}};
    const codeView = document.querySelector('#code-view');
    const confirmView = document.querySelector('#confirm-view');
    const statusView = document.querySelector('#status-view');
    const codeInput = document.querySelector('#code');
    let candidateId = '';
    let pollTimer;

    const post = async (path, body = {}) => {
      const response = await fetch(base + path, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(body),
        cache: 'no-store',
        credentials: 'omit'
      });
      const value = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(value?.message || 'İşlem tamamlanamadı.');
      return value;
    };

    const activateStep = (number) => {
      ['code', 'confirm', 'login', 'done'].forEach((name, index) => {
        const element = document.querySelector('#step-' + name);
        element.classList.toggle('active', index === number - 1);
        element.classList.toggle('complete', index < number - 1);
      });
    };

    const showError = (message) => {
      codeView.hidden = true;
      confirmView.hidden = true;
      statusView.hidden = false;
      document.querySelector('#status-card').dataset.phase = 'error';
      document.querySelector('#status-title').textContent = 'Bağlantı kurulamadı';
      document.querySelector('#status-message').textContent = message;
      document.querySelector('#retry').hidden = false;
      activateStep(1);
    };

    document.querySelector('#paste').addEventListener('click', async () => {
      try {
        codeInput.value = await navigator.clipboard.readText();
      } catch {
        codeInput.focus();
      }
    });

    document.querySelector('#inspect').addEventListener('click', async () => {
      try {
        const status = await post('inspect', {code: codeInput.value});
        codeInput.value = '';
        candidateId = status.candidateId;
        document.querySelector('#gateway-host').textContent = status.gatewayHost;
        document.querySelector('#expiry').textContent = new Date(status.expiresAt).toLocaleString();
        codeView.hidden = true;
        confirmView.hidden = false;
        statusView.hidden = true;
        activateStep(2);
      } catch (error) {
        showError(error.message);
      }
    });

    document.querySelector('#back').addEventListener('click', async () => {
      await post('reset');
      candidateId = '';
      codeView.hidden = false;
      confirmView.hidden = true;
      statusView.hidden = true;
      codeInput.focus();
      activateStep(1);
    });

    document.querySelector('#connect').addEventListener('click', async () => {
      try {
        await post('connect', {candidateId});
        confirmView.hidden = true;
        statusView.hidden = false;
        document.querySelector('#status-card').dataset.phase = 'connecting';
        document.querySelector('#status-title').textContent = 'DeepSeek girişi bekleniyor';
        document.querySelector('#status-message').textContent = 'Açılan tarayıcı penceresinde giriş ve doğrulama adımlarını tamamlayın.';
        activateStep(3);
        pollTimer = window.setInterval(refreshStatus, 900);
      } catch (error) {
        showError(error.message);
      }
    });

    const refreshStatus = async () => {
      try {
        const response = await fetch(base + 'status', {cache: 'no-store', credentials: 'omit'});
        const status = await response.json();
        if (status.phase === 'connecting') return;
        window.clearInterval(pollTimer);
        document.querySelector('#status-card').dataset.phase = status.phase;
        document.querySelector('#status-message').textContent = status.message;
        if (status.phase === 'complete') {
          document.querySelector('#status-title').textContent = 'Hesap bağlandı';
          document.querySelector('#close').hidden = false;
          activateStep(4);
        } else {
          document.querySelector('#status-title').textContent = 'Bağlantı kurulamadı';
          document.querySelector('#retry').hidden = false;
          activateStep(3);
        }
      } catch {
        window.clearInterval(pollTimer);
      }
    };

    document.querySelector('#retry').addEventListener('click', async () => {
      await post('reset');
      document.querySelector('#retry').hidden = true;
      statusView.hidden = true;
      codeView.hidden = false;
      codeInput.focus();
      activateStep(1);
    });

    document.querySelector('#close').addEventListener('click', async () => {
      await post('shutdown').catch(() => undefined);
      window.close();
      document.querySelector('#status-message').textContent = 'Connector kapatıldı. Bu sekmeyi kapatabilirsiniz.';
    });
  </script>
</body>
</html>`
