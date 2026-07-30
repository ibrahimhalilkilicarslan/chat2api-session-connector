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
    .status-hint {
      margin-top: 10px !important; padding-top: 10px; border-top: 1px solid var(--line);
      color: var(--text) !important;
    }
    .status-code {
      display: inline-flex; margin-top: 10px; padding: 4px 7px;
      border: 1px solid var(--line); border-radius: 6px;
      color: var(--muted); background: rgba(0, 0, 0, .18);
      font: 10px/1.2 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    }
    .privacy { margin: 22px 0 0; padding-top: 18px; border-top: 1px solid var(--line); color: #708981; font-size: 11px; line-height: 1.65; }
    .notice {
      margin: 0 0 18px; padding: 11px 13px;
      border: 1px solid rgba(255, 196, 105, .2); border-radius: 11px;
      color: #f5d59e; background: rgba(255, 196, 105, .06);
      font-size: 11px; line-height: 1.55;
    }
    .install-success { text-align: center; }
    .success-mark {
      width: 58px; height: 58px; display: grid; place-items: center; margin: 0 auto 17px;
      border: 1px solid rgba(68, 215, 170, .35); border-radius: 18px;
      color: #04110d; background: linear-gradient(145deg, var(--accent-2), var(--accent));
      font-size: 27px; font-weight: 900; box-shadow: 0 16px 38px rgba(68, 215, 170, .14);
    }
    .install-success h2 { margin: 0; font-size: 22px; letter-spacing: -.025em; }
    .install-success > p { max-width: 510px; margin: 10px auto 0; color: var(--muted); font-size: 13px; line-height: 1.65; }
    .guide {
      margin: 23px 0 0; padding: 0; display: grid; gap: 9px;
      list-style: none; text-align: left; counter-reset: guide;
    }
    .guide li {
      min-height: 52px; display: flex; align-items: center; gap: 12px; padding: 11px 13px;
      border: 1px solid var(--line); border-radius: 13px; background: rgba(4, 12, 10, .42);
      color: var(--muted); font-size: 12px; line-height: 1.45; counter-increment: guide;
    }
    .guide li::before {
      content: counter(guide); width: 28px; height: 28px; display: grid; place-items: center; flex: 0 0 auto;
      border: 1px solid rgba(68, 215, 170, .3); border-radius: 9px; color: var(--accent);
      background: rgba(68, 215, 170, .07); font-weight: 850;
    }
    .guide strong { color: var(--text); }
    .install-actions { justify-content: center; margin-top: 18px; }
    .version { display: block; margin-top: 15px; color: #617a72; font-size: 10px; }
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
        {{if .StandaloneLaunch}}
          {{if .InstallationReady}}
        <h1>Connector kullanıma hazır</h1>
        <p>Kurulum tamamlandı. Hesap bağlantısını Chat2API panelinden başlattığınızda connector gerekli adımları otomatik olarak açacak.</p>
          {{else}}
        <h1>Connector kurulumu tamamlanamadı</h1>
        <p>Bağlantı protokolü kaydedilemedi. Aşağıdaki onarım adımlarını uygulayın; bağlantı kodu yalnız yedek yöntemdir.</p>
          {{end}}
        {{else}}
        <h1>DeepSeek hesabınızı bağlayın</h1>
        <p>Giriş doğrudan DeepSeek üzerinde tamamlanır. Connector parola veya doğrulama kodunuzu görmez; yalnız doğrulanmış oturum bilgisini seçtiğiniz Chat2API gateway’ine iletir.</p>
        {{end}}
      </div>
      <div class="content">
        {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
        <section class="install-success" id="install-view" {{if .StandaloneLaunch}}{{else}}hidden{{end}}>
          <div class="success-mark" aria-hidden="true">{{if .InstallationReady}}✓{{else}}!{{end}}</div>
          {{if .InstallationReady}}
          <h2>Kurulum tamamlandı</h2>
          <p>Bu uygulamada kod aramanıza gerek yok. Bağlantı işlemini Chat2API yönetim panelinden başlatın.</p>
          <ol class="guide">
            <li><span><strong>Chat2API paneline dönün.</strong> Açık connector sekmesini kapatabilirsiniz.</span></li>
            <li><span><strong>DeepSeek hesabı ekle</strong> alanından “Connector ile bağlan” seçeneğini kullanın.</span></li>
            <li><span>Açılan güvenli DeepSeek penceresinde girişinizi tamamlayın. Oturum otomatik olarak hesabınıza bağlanır.</span></li>
          </ol>
          {{else}}
          <h2>Bağlantı protokolünü onarın</h2>
          <p>Connector’ın başka bir kopyası açıksa kapatın ve indirdiğiniz uygulamayı bir kez daha çalıştırın.</p>
          <ol class="guide">
            <li><span>Açık <strong>Chat2API Connector</strong> pencerelerinin tamamını kapatın.</span></li>
            <li><span>İndirdiğiniz connector uygulamasını yeniden çalıştırın ve “Kurulum tamamlandı” ekranını bekleyin.</span></li>
            <li><span>Chat2API paneline dönüp <strong>Connector ile bağlan</strong> seçeneğini yeniden kullanın.</span></li>
          </ol>
          {{end}}
          <div class="actions install-actions">
            <button id="finish-install" type="button">{{if .InstallationReady}}Tamam, Chat2API’ye dön{{else}}Connector’ı kapat{{end}}</button>
            <button id="manual-code" type="button" class="secondary">Manuel bağlantı kodum var</button>
          </div>
          {{if .Version}}<small class="version">Connector {{.Version}}</small>{{end}}
        </section>

        <div id="connection-flow" {{if .StandaloneLaunch}}hidden{{end}}>
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
              <div>
                <strong id="status-title"></strong>
                <p id="status-message"></p>
                <p class="status-hint" id="status-hint" hidden></p>
                <code class="status-code" id="status-code" hidden></code>
              </div>
            </div>
            <div class="actions" id="status-actions">
              <button id="retry" type="button" class="secondary" hidden>Tekrar dene</button>
              <button id="close" type="button" class="ghost" hidden>Connector’ı kapat</button>
            </div>
          </section>
        </div>

        <p class="privacy">Kişisel tarayıcı profiliniz kullanılmaz. Geçici profil işlem sonunda silinir; token diske veya uygulama loglarına yazılmaz.</p>
      </div>
    </section>
  </main>
  <script nonce="{{.Nonce}}">
    const base = {{.BasePath}};
    const installView = document.querySelector('#install-view');
    const connectionFlow = document.querySelector('#connection-flow');
    const standaloneLaunch = !installView.hidden;
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
      if (!response.ok) {
        const error = new Error(value?.message || 'İşlem tamamlanamadı.');
        error.hint = value?.hint || '';
        error.errorCode = value?.errorCode || '';
        throw error;
      }
      return value;
    };

    const activateStep = (number) => {
      ['code', 'confirm', 'login', 'done'].forEach((name, index) => {
        const element = document.querySelector('#step-' + name);
        element.classList.toggle('active', index === number - 1);
        element.classList.toggle('complete', index < number - 1);
      });
    };

    const showConnectionFlow = () => {
      installView.hidden = true;
      connectionFlow.hidden = false;
    };

    const showError = (message, hint = '', errorCode = '') => {
      showConnectionFlow();
      codeView.hidden = true;
      confirmView.hidden = true;
      statusView.hidden = false;
      document.querySelector('#status-card').dataset.phase = 'error';
      document.querySelector('#status-title').textContent = 'Bağlantı kurulamadı';
      document.querySelector('#status-message').textContent = message;
      document.querySelector('#status-hint').textContent = hint;
      document.querySelector('#status-hint').hidden = !hint;
      document.querySelector('#status-code').textContent = errorCode;
      document.querySelector('#status-code').hidden = !errorCode;
      document.querySelector('#retry').hidden = false;
      activateStep(1);
    };

    const showConfirm = (status) => {
      showConnectionFlow();
      candidateId = status.candidateId;
      document.querySelector('#gateway-host').textContent = status.gatewayHost;
      document.querySelector('#expiry').textContent = new Date(status.expiresAt).toLocaleString();
      codeView.hidden = true;
      confirmView.hidden = false;
      statusView.hidden = true;
      activateStep(2);
    };

    document.querySelector('#manual-code').addEventListener('click', () => {
      showConnectionFlow();
      codeInput.focus();
    });

    document.querySelector('#finish-install').addEventListener('click', async (event) => {
      event.currentTarget.disabled = true;
      event.currentTarget.textContent = 'Chat2API paneline dönebilirsiniz';
      await post('shutdown').catch(() => undefined);
      window.close();
    });

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
        showConfirm(status);
      } catch (error) {
        showError(error.message, error.hint, error.errorCode);
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
        document.querySelector('#status-hint').hidden = true;
        document.querySelector('#status-code').hidden = true;
        activateStep(3);
        pollTimer = window.setInterval(refreshStatus, 900);
      } catch (error) {
        showError(error.message, error.hint, error.errorCode);
      }
    });

    const refreshStatus = async () => {
      try {
        const response = await fetch(base + 'status', {cache: 'no-store', credentials: 'omit'});
        const status = await response.json();
        if (status.phase === 'connecting') {
          document.querySelector('#status-message').textContent = status.message;
          return;
        }
        window.clearInterval(pollTimer);
        document.querySelector('#status-card').dataset.phase = status.phase;
        document.querySelector('#status-message').textContent = status.message;
        if (status.phase === 'complete') {
          document.querySelector('#status-title').textContent = 'Hesap bağlandı';
          document.querySelector('#close').hidden = false;
          activateStep(4);
        } else {
          document.querySelector('#status-title').textContent = 'Bağlantı kurulamadı';
          document.querySelector('#status-hint').textContent = status.hint || '';
          document.querySelector('#status-hint').hidden = !status.hint;
          document.querySelector('#status-code').textContent = status.errorCode || '';
          document.querySelector('#status-code').hidden = !status.errorCode;
          document.querySelector('#retry').hidden = false;
          activateStep(3);
        }
      } catch {
        window.clearInterval(pollTimer);
      }
    };

    document.querySelector('#retry').addEventListener('click', async () => {
      try {
        const response = await fetch(base + 'status', {cache: 'no-store', credentials: 'omit'});
        const status = await response.json().catch(() => ({}));
        if (status.candidateId) {
          document.querySelector('#retry').hidden = true;
          document.querySelector('#status-hint').hidden = true;
          document.querySelector('#status-code').hidden = true;
          showConfirm(status);
          return;
        }
        await post('reset');
        document.querySelector('#retry').hidden = true;
        statusView.hidden = true;
        codeView.hidden = false;
        codeInput.focus();
        activateStep(1);
      } catch (error) {
        showError(error.message, error.hint, error.errorCode);
      }
    });

    document.querySelector('#close').addEventListener('click', async () => {
      await post('shutdown').catch(() => undefined);
      window.close();
      document.querySelector('#status-message').textContent = 'Connector kapatıldı. Bu sekmeyi kapatabilirsiniz.';
    });

    const hydrate = async () => {
      try {
        const response = await fetch(base + 'status', {cache: 'no-store', credentials: 'omit'});
        const status = await response.json();
        if (status.phase === 'confirm') {
          showConfirm(status);
        } else if (status.phase === 'error') {
          showError(status.message, status.hint, status.errorCode);
        }
      } catch {
        showError(
          'Connector yerel durumunu okuyamadı.',
          'Connector uygulamasını kapatıp yeniden açın.',
          'LOCAL_STATUS_UNAVAILABLE'
        );
      }
    };

    if (!standaloneLaunch) {
      void hydrate();
    }
  </script>
</body>
</html>`
