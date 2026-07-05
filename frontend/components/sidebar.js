function renderSidebar(activeView) {
    const sidebarContainer = document.getElementById('sidebar-container');
    if (!sidebarContainer) return;

    const userJson = localStorage.getItem('current_user');
    const user = userJson ? JSON.parse(userJson) : null;
    const isAdmin = user ? user.is_admin : false;

    let adminLinks = '';
    if (isAdmin) {
        adminLinks = `
            <a href="/pages/users.html" class="nav-item ${activeView === 'users' ? 'active' : ''}">
                <i class="fa-solid fa-users-gear"></i>
                <span>User Management</span>
            </a>
            <a href="/pages/groups.html" class="nav-item ${activeView === 'groups' ? 'active' : ''}">
                <i class="fa-solid fa-users"></i>
                <span>Group Management</span>
            </a>
            <a href="/pages/roles.html" class="nav-item ${activeView === 'roles' ? 'active' : ''}">
                <i class="fa-solid fa-shield-halved"></i>
                <span>Role Management</span>
            </a>
        `;
    }

    sidebarContainer.innerHTML = `
        <aside class="app-sidebar">
            <nav class="sidebar-nav">
                <a href="/pages/dashboard.html" class="nav-item ${activeView === 'dashboard' ? 'active' : ''}">
                    <i class="fa-solid fa-user"></i>
                    <span>Profile Dashboard</span>
                </a>
                
                ${adminLinks}
                
                <button id="logout-sidebar-btn" class="nav-item logout-item">
                    <i class="fa-solid fa-arrow-right-from-bracket"></i>
                    <span>Log Out</span>
                </button>
            </nav>
        </aside>
    `;

    document.getElementById('logout-sidebar-btn').addEventListener('click', async () => {
        try {
            await fetch('/api/auth/logout', {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${localStorage.getItem('jwt_token')}` }
            });
        } catch (e) {}
        localStorage.removeItem('jwt_token');
        localStorage.removeItem('current_user');
        window.location.href = '/pages/login.html';
    });
}
