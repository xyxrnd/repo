<script>
  import Sidebar from "../components/common/Sidebar.svelte";
  import Topbar from "../components/common/Topbar.svelte";

  let sidebarOpen = false;

  function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
  }

  function closeSidebar() {
    sidebarOpen = false;
  }
</script>

<div class="flex h-screen overflow-hidden">
  <!-- Mobile Overlay -->
  {#if sidebarOpen}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
      class="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm lg:hidden"
      on:click={closeSidebar}
    ></div>
  {/if}

  <!-- Sidebar: hidden on mobile, visible on lg+ -->
  <div
    class="fixed inset-y-0 left-0 z-50 transform transition-transform duration-300 ease-in-out lg:relative lg:translate-x-0 {sidebarOpen
      ? 'translate-x-0'
      : '-translate-x-full'}"
  >
    <Sidebar on:navigate={closeSidebar} />
  </div>

  <main class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <Topbar on:toggleSidebar={toggleSidebar} />

    <div
      class="flex-1 overflow-y-auto p-4 md:p-6 lg:p-8 bg-background-light dark:bg-background-dark/50"
    >
      <slot />
    </div>
  </main>
</div>
