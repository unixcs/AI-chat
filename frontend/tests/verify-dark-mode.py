from playwright.sync_api import sync_playwright


USER_PHONE = "15555166986"
USER_PASSWORD = "simple1270."
ADMIN_USERNAME = "admin"
ADMIN_PASSWORD = "admin123"
ADMIN_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbklkIjoidXNlcl9hZG1pbiIsInR5cGUiOiJhZG1pbiIsImlhdCI6MTc3NjM1MzgxNywiZXhwIjoxNzc2OTU4NjE3fQ.LMJ9P6MnhK9dsyND5J84C8RwEJIGI-JCQqCJpT6n_0A"


def assert_dark_surface(page, selector, description):
    bg = page.locator(selector).first.evaluate(
        "el => getComputedStyle(el).backgroundImage + ' | ' + getComputedStyle(el).backgroundColor"
    )
    print(f"{description}: {bg}")
    assert "rgb(255, 255, 255)" not in bg, (
        f"{description} still renders a bright white surface: {bg}"
    )


with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)

    user_page = browser.new_page()
    user_page.goto("http://localhost:5173/login")
    user_page.wait_for_load_state("networkidle")
    user_page.get_by_placeholder("请输入手机号").fill(USER_PHONE)
    user_page.get_by_placeholder("请输入密码").fill(USER_PASSWORD)
    user_page.get_by_role("button", name="登录").click()
    user_page.wait_for_url("**/app/chat")
    user_page.wait_for_load_state("networkidle")
    user_page.evaluate(
        "() => { document.documentElement.dataset.theme = 'dark'; localStorage.setItem('theme', 'dark') }"
    )
    user_page.goto("http://localhost:5173/app/profile")
    user_page.wait_for_load_state("networkidle")
    assert_dark_surface(user_page, ".accentPanel", "Profile membership card")
    assert_dark_surface(user_page, ".menuBtn.active", "User sidebar active nav")
    user_page.goto("http://localhost:5173/app/chat")
    user_page.wait_for_load_state("networkidle")
    assert_dark_surface(user_page, ".composerSurface", "Chat composer surface")
    user_page.locator(".historyIconBtn").click()
    user_page.wait_for_timeout(200)
    assert_dark_surface(
        user_page, ".historyDrawerPanel .historyItem", "Chat history drawer item"
    )

    admin_page = browser.new_page()
    admin_page.goto("http://localhost:5173/admin/login")
    admin_page.wait_for_load_state("networkidle")
    admin_page.evaluate(
        f"() => {{ localStorage.setItem('adminToken', '{ADMIN_TOKEN}'); document.documentElement.dataset.theme = 'dark'; localStorage.setItem('theme', 'dark') }}"
    )
    admin_page.goto("http://localhost:5173/admin/conversations")
    admin_page.wait_for_load_state("networkidle")
    assert_dark_surface(admin_page, ".adminSidebar", "Admin sidebar")
    assert_dark_surface(admin_page, ".menuGroup button.active", "Admin active nav item")
    assert_dark_surface(admin_page, ".queryCluster", "Admin conversation query cluster")
    assert_dark_surface(admin_page, ".historyItem", "Admin conversation list item")
    assert_dark_surface(
        admin_page, ".transcriptBubble.assistant", "Admin transcript assistant bubble"
    )

    browser.close()
