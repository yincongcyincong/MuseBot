// Change-password prompt for accounts still using the default admin/admin
// password. Self-contained plain JS: the SPA bundle is compiled without
// sources in this repo, so this overlay is injected via a script tag in
// index.html instead. The prompt can be dismissed ("稍后再说") and stays
// dismissed for the rest of the browser session (sessionStorage flag).
(function () {
  'use strict';

  var DISMISS_KEY = 'musebot_default_pwd_dismissed';
  var POLL_INTERVAL_MS = 1500;
  var MAX_POLLS = 20;

  function dismissed() {
    try {
      return sessionStorage.getItem(DISMISS_KEY) === '1';
    } catch (e) {
      return false;
    }
  }

  function markDismissed() {
    try {
      sessionStorage.setItem(DISMISS_KEY, '1');
    } catch (e) {
      /* private mode etc. */
    }
  }

  function fetchJSON(url, opts) {
    opts = opts || {};
    opts.credentials = 'same-origin';
    return fetch(url, opts).then(function (res) {
      return res.json();
    });
  }

  function el(tag, style, text) {
    var node = document.createElement(tag);
    if (style) node.style.cssText = style;
    if (text != null) node.textContent = text;
    return node;
  }

  function button(label, background, color) {
    var btn = el('button', 'padding:6px 14px;border:none;border-radius:4px;font-size:13px;cursor:pointer;background:' + background + ';color:' + color + ';', label);
    return btn;
  }

  function input(placeholder) {
    var field = el('input');
    field.type = 'password';
    field.placeholder = placeholder;
    field.style.cssText = 'width:100%;box-sizing:border-box;padding:7px 9px;border:1px solid #d0d0d0;border-radius:4px;font-size:13px;margin-bottom:10px;';
    return field;
  }

  function openModal(userId) {
    var overlay = el('div', 'position:fixed;inset:0;background:rgba(0,0,0,.45);display:flex;align-items:center;justify-content:center;z-index:99999;');
    var card = el('div', 'background:#fff;border-radius:8px;padding:24px;width:320px;box-shadow:0 8px 30px rgba(0,0,0,.25);font-family:system-ui,-apple-system,sans-serif;color:#333;position:relative;');

    var close = button('×', 'transparent', '#999');
    close.style.cssText += ';position:absolute;top:6px;right:10px;font-size:18px;line-height:1;';
    var title = el('div', 'font-size:16px;font-weight:600;margin-bottom:8px;', '修改默认密码');
    var desc = el('div', 'font-size:13px;color:#666;margin-bottom:14px;line-height:1.5;', '当前账户仍在使用默认密码 admin，建议尽快修改。');
    var pwd = input('新密码（至少 6 位）');
    var confirmPwd = input('确认新密码');
    var msg = el('div', 'font-size:12px;color:#e5484d;min-height:16px;margin:4px 0;');
    var row = el('div', 'display:flex;justify-content:flex-end;gap:8px;margin-top:8px;');
    var later = button('稍后再说', '#f0f0f0', '#333');
    var ok = button('确认修改', '#4f6ef2', '#fff');

    function closeAndDismiss() {
      markDismissed();
      if (overlay.parentNode) document.body.removeChild(overlay);
      document.removeEventListener('keydown', onKey);
    }

    function onKey(e) {
      if (e.key === 'Escape') closeAndDismiss();
    }

    function submit() {
      var next = pwd.value;
      if (!next || next.length < 6) {
        msg.textContent = '新密码至少需要 6 位';
        return;
      }
      if (next !== confirmPwd.value) {
        msg.textContent = '两次输入的密码不一致';
        return;
      }
      msg.style.color = '#999';
      msg.textContent = '提交中…';
      ok.disabled = true;
      fetchJSON('/user/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: userId, password: next })
      }).then(function (res) {
        if (res && res.code === 0) {
          msg.style.color = '#2da44e';
          msg.textContent = '密码修改成功';
          setTimeout(closeAndDismiss, 1200);
        } else {
          ok.disabled = false;
          msg.style.color = '#e5484d';
          msg.textContent = (res && res.message) || '修改失败，请重试';
        }
      }).catch(function () {
        ok.disabled = false;
        msg.style.color = '#e5484d';
        msg.textContent = '网络错误，请重试';
      });
    }

    close.addEventListener('click', closeAndDismiss);
    later.addEventListener('click', closeAndDismiss);
    ok.addEventListener('click', submit);
    document.addEventListener('keydown', onKey);

    row.appendChild(later);
    row.appendChild(ok);
    card.appendChild(close);
    card.appendChild(title);
    card.appendChild(desc);
    card.appendChild(pwd);
    card.appendChild(confirmPwd);
    card.appendChild(msg);
    card.appendChild(row);
    overlay.appendChild(card);
    document.body.appendChild(overlay);
    pwd.focus();
  }

  function poll(remaining) {
    if (remaining <= 0 || dismissed()) return;
    fetchJSON('/user/me').then(function (res) {
      if (!res || res.code !== 0) {
        // not logged in yet (login page) — keep waiting
        setTimeout(function () { poll(remaining - 1); }, POLL_INTERVAL_MS);
        return;
      }
      var data = res.data || {};
      if (data.is_default_password && !dismissed()) {
        openModal(data.user_id);
      }
    }).catch(function () {
      setTimeout(function () { poll(remaining - 1); }, POLL_INTERVAL_MS);
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { poll(MAX_POLLS); });
  } else {
    poll(MAX_POLLS);
  }
})();
