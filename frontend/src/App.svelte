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

  // Auth Pages
  import LoginPage from "./pages/Auth/LoginPage.svelte";
  import RegisterPage from "./pages/Auth/RegisterPage.svelte";

  // Public Pages
  import LandingPage from "./pages/Landing/LandingPage.svelte";
  import BrowsePage from "./pages/Browse/BrowsePage.svelte";
  import AboutPage from "./pages/About/AboutPage.svelte";
  import DocumentDetail from "./pages/Documents/DocumentDetail.svelte";
  import NotFound from "./pages/NotFound.svelte";

  // Track current route for layout switching
  let currentRoute = "/";

  // Auth routes (no layout wrapper, they have their own full-page layout)
  const authRoutes = ["/login", "/register"];

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

  // Wrapper for protected routes
  function protectedRoute(component) {
    return wrap({
      component,
      conditions: [requireAuth],
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

    // Admin routes (protected)
    "/admin": protectedRoute(Dashboard),
    "/admin/dashboard": protectedRoute(Dashboard),
    "/admin/documents": protectedRoute(DocumentList),
    "/admin/documents/add": protectedRoute(DocumentAdd),
    "/admin/documents/edit/:id": protectedRoute(DocumentEdit),
    "/admin/users": protectedRoute(UserList),
    "/admin/fakultas": protectedRoute(FakultasList),
    "/admin/prodi": protectedRoute(ProdiList),
    "/dashboard": protectedRoute(Dashboard),
    "/documents": protectedRoute(DocumentList),
    "/documents/add": protectedRoute(DocumentAdd),
    "/documents/edit/:id": protectedRoute(DocumentEdit),
    "/users": protectedRoute(UserList),
    "/fakultas": protectedRoute(FakultasList),
    "/prodi": protectedRoute(ProdiList),
    "/reports": protectedRoute(Reports),
    "/settings": protectedRoute(Settings),
    "/system-settings": protectedRoute(SystemSettings),
    "/admin/system-settings": protectedRoute(SystemSettings),
    "/access-requests": protectedRoute(AccessRequestList),
    "/admin/access-requests": protectedRoute(AccessRequestList),

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
