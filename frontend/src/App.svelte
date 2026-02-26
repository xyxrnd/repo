<script>
  import Router from "svelte-spa-router";
  import { wrap } from "svelte-spa-router/wrap";
  import { push } from "svelte-spa-router";
  import { authService } from "./services/authService.js";

  // Layouts
  import AdminLayout from "./layouts/AdminLayout.svelte";
  import PublicLayout from "./layouts/PublicLayout.svelte";

  // Admin Pages
  import Dashboard from "./pages/Admin/Dashboard.svelte";
  import Reports from "./pages/Admin/Reports.svelte";
  import Settings from "./pages/Admin/Settings.svelte";
  import DocumentList from "./pages/Documents/DocumentList.svelte";
  import DocumentAdd from "./pages/Documents/DocumentAdd.svelte";
  import DocumentEdit from "./pages/Documents/DocumentEdit.svelte";
  import UserList from "./pages/Users/UserList.svelte";
  import FakultasList from "./pages/Fakultas/FakultasList.svelte";
  import ProdiList from "./pages/Prodi/ProdiList.svelte";
  import SystemSettings from "./pages/Admin/SystemSettings.svelte";
  import AccessRequestList from "./pages/Admin/AccessRequestList.svelte";
  import StudentRegistrationList from "./pages/Admin/StudentRegistrationList.svelte";

  // Auth Pages
  import LoginPage from "./pages/Auth/LoginPage.svelte";
  import RegisterPage from "./pages/Auth/RegisterPage.svelte";
  import StudentSignupPage from "./pages/Auth/StudentSignupPage.svelte";

  // Public Pages
  import LandingPage from "./pages/Landing/LandingPage.svelte";
  import BrowsePage from "./pages/Browse/BrowsePage.svelte";
  import AboutPage from "./pages/About/AboutPage.svelte";
  import DocumentDetail from "./pages/Documents/DocumentDetail.svelte";
  import NotFound from "./pages/NotFound.svelte";

  // Track current route for layout switching
  let currentRoute = "/";

  // Auth routes (no layout wrapper, they have their own full-page layout)
  const authRoutes = ["/login", "/register", "/student-signup"];

  // Public routes that use PublicLayout
  const publicRoutes = ["/", "/landing", "/browse", "/about", "/document"];

  function isAuthRoute(route) {
    return authRoutes.some((r) => route === r || route.startsWith(r + "/"));
  }

  function isPublicRoute(route) {
    return publicRoutes.some((r) => route === r || route.startsWith(r + "/"));
  }

  // Auth guard function for protected routes
  function requireAuth() {
    if (!authService.isAuthenticated()) {
      // Store the intended destination for redirect after login
      sessionStorage.setItem("redirectAfterLogin", window.location.hash);
      push("/login");
      return false;
    }
    return true;
  }

  // Admin guard function - checks both auth and admin role
  function requireAdmin() {
    if (!authService.isAuthenticated()) {
      sessionStorage.setItem("redirectAfterLogin", window.location.hash);
      push("/login");
      return false;
    }
    if (!authService.isAdmin()) {
      // Mahasiswa or non-admin users are redirected to home
      push("/");
      return false;
    }
    return true;
  }

  // Wrapper for protected routes (any authenticated user)
  function protectedRoute(component) {
    return wrap({
      component,
      conditions: [requireAuth],
    });
  }

  // Wrapper for admin-only routes
  function adminRoute(component) {
    return wrap({
      component,
      conditions: [requireAdmin],
    });
  }

  // Combined routes
  const routes = {
    // Landing / Public routes
    "/": LandingPage,
    "/landing": LandingPage,
    "/browse": BrowsePage,
    "/about": AboutPage,
    "/document/:id": DocumentDetail,

    // Auth routes
    "/login": LoginPage,
    "/register": RegisterPage,
    "/student-signup": StudentSignupPage,

    // Admin routes (admin only - mahasiswa cannot access)
    "/admin": adminRoute(Dashboard),
    "/admin/dashboard": adminRoute(Dashboard),
    "/admin/documents": adminRoute(DocumentList),
    "/admin/documents/add": adminRoute(DocumentAdd),
    "/admin/documents/edit/:id": adminRoute(DocumentEdit),
    "/admin/users": adminRoute(UserList),
    "/admin/fakultas": adminRoute(FakultasList),
    "/admin/prodi": adminRoute(ProdiList),
    "/dashboard": adminRoute(Dashboard),
    "/documents": adminRoute(DocumentList),
    "/documents/add": adminRoute(DocumentAdd),
    "/documents/edit/:id": adminRoute(DocumentEdit),
    "/users": adminRoute(UserList),
    "/fakultas": adminRoute(FakultasList),
    "/prodi": adminRoute(ProdiList),
    "/reports": adminRoute(Reports),
    "/settings": adminRoute(Settings),
    "/system-settings": adminRoute(SystemSettings),
    "/admin/system-settings": adminRoute(SystemSettings),
    "/access-requests": adminRoute(AccessRequestList),
    "/admin/access-requests": adminRoute(AccessRequestList),
    "/student-registrations": adminRoute(StudentRegistrationList),
    "/admin/student-registrations": adminRoute(StudentRegistrationList),

    // Catch-all
    "*": NotFound,
  };

  function routeLoaded(event) {
    currentRoute = event.detail.location;
  }
</script>

{#if isAuthRoute(currentRoute)}
  <!-- Auth pages have their own full-page layout -->
  <Router {routes} on:routeLoaded={routeLoaded} />
{:else if isPublicRoute(currentRoute)}
  <PublicLayout>
    <Router {routes} on:routeLoaded={routeLoaded} />
  </PublicLayout>
{:else}
  <AdminLayout>
    <Router {routes} on:routeLoaded={routeLoaded} />
  </AdminLayout>
{/if}
