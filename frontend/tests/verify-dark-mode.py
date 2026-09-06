"""暗色模式验证（Part D 新体系）：双机制并存 + 暗色下无刺眼白面。

前置：dev server（npm run dev，5173）+ Go 后端（3001）可用；
USER_PHONE / USER_PASSWORD / ADMIN_USERNAME / ADMIN_PASSWORD 由环境变量提供
（对齐 yun1 测试实例凭据；不再硬编码任何令牌）。
"""

import os

from playwright.sync_api import sync_playwright

USER_PHONE = os.environ["USER_PHONE"]
USER_PASSWORD = os.environ["USER_PASSWORD"]
ADMIN_USERNAME = os.environ.get("ADMIN_USERNAME", "admin")
ADMIN_PASSWORD = os.environ["ADMIN_PASSWORD"]


def assert_dark_surface(page, selector, description):
    bg = page.locator(selector).first.evaluate(
        "el => getComputedStyle(el).backgroundImage + ' | ' + getComputedStyle(el).backgroundColor"
    )
    print(f"{description}: {bg}")
    assert "rgb(255, 255, 255)" not in bg, (
        f"{description} still renders a bright white surface: {bg}"
    )


def assert_dual_mechanism(page, description):
    """theme.js 必须同时驱动 data-theme 属性与 html.dark class（Plan §3.4）。"""
    state = page.evaluate(
        "() => ({"
        " theme: document.documentElement.dataset.theme,"
        " darkClass: document.documentElement.classList.contains('dark'),"
        " bgToken: getComputedStyle(document.documentElement).getPropertyValue('--background').trim()"
        "})"
    )
    print(f"{description}: {state}")
    assert state["theme"] == "dark", f"{description}: data-theme not dark: {state}"
    assert state["darkClass"] is True, f"{description}: html.dark class missing (dual mechanism broken): {state}"
    # 暗色下 --background 必须是深雾底（亮色为 #f5f4ef）
    assert state["bgToken"] not in ("#f5f4ef", "#F5F4EF"), (
        f"{description}: --background still resolves to the light token: {state}"
    )


def set_dark_and_reload(page, url):
    page.evaluate("() => { localStorage.setItem('theme', 'dark') }")
    page.goto(url)
    page.wait_for_load_state("networkidle")


with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)

    # ---------- 用户端 ----------
    user_page = browser.new_page()
    user_page.goto("http://localhost:5173/login")
    user_page.wait_for_load_state("networkidle")
    user_page.get_by_placeholder("请输入手机号").fill(USER_PHONE)
    user_page.get_by_placeholder("请输入密码").fill(USER_PASSWORD)
    user_page.get_by_role("button", name="登录").click()
    user_page.wait_for_url("**/app/chat")
    user_page.wait_for_load_state("networkidle")

    set_dark_and_reload(user_page, "http://localhost:5173/app/chat")
    assert_dual_mechanism(user_page, "User chat (dual mechanism)")

    # 消息气泡与侧栏在暗色下不刺眼
    if user_page.locator(".chatBotBubble").count() > 0:
        assert_dark_surface(user_page, ".chatBotBubble", "Chat bot bubble")
    user_page.goto("http://localhost:5173/app/profile")
    user_page.wait_for_load_state("networkidle")
    assert_dark_surface(user_page, "aside", "User sidebar")
    assert_dark_surface(user_page, "main .bg-card", "Profile card")

    # ---------- 管理端 ----------
    admin_page = browser.new_page()
    admin_page.goto("http://localhost:5173/admin/login")
    admin_page.wait_for_load_state("networkidle")
    admin_page.get_by_placeholder("请输入管理员账号").fill(ADMIN_USERNAME)
    admin_page.get_by_placeholder("请输入密码").fill(ADMIN_PASSWORD)
    admin_page.get_by_role("button", name="登录后台").click()
    admin_page.wait_for_url("**/admin")
    admin_page.wait_for_load_state("networkidle")

    set_dark_and_reload(admin_page, "http://localhost:5173/admin/conversations")
    assert_dual_mechanism(admin_page, "Admin console (dual mechanism)")
    assert_dark_surface(admin_page, "aside", "Admin sidebar")
    assert_dark_surface(admin_page, "main .bg-card", "Admin conversation card")

    # 切回浅色：双机制同步撤销
    admin_page.evaluate("() => { localStorage.setItem('theme', 'light') }")
    admin_page.goto("http://localhost:5173/admin")
    admin_page.wait_for_load_state("networkidle")
    light_state = admin_page.evaluate(
        "() => ({"
        " theme: document.documentElement.dataset.theme,"
        " darkClass: document.documentElement.classList.contains('dark')"
        "})"
    )
    print(f"Admin light toggle: {light_state}")
    assert light_state["theme"] == "light" and light_state["darkClass"] is False, (
        f"light toggle must clear both mechanisms: {light_state}"
    )

    browser.close()

print("DARK MODE OK: dual mechanism asserted, no bright surfaces in dark")
