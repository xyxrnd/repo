<script>
    import { onMount } from "svelte";
    import userService from "../../services/userService";
    import authService from "../../services/authService";

    let users = [];
    let loading = true;
    let error = "";
    let showModal = false;
    let editingUser = null;
    let deleteConfirm = null;

    // Form fields
    let formName = "";
    let formEmail = "";
    let formPassword = "";
    let formRole = "mahasiswa";
    let formError = "";
    let formLoading = false;

    const currentUser = authService.getUser();

    onMount(async () => {
        await loadUsers();
    });

    async function loadUsers() {
        loading = true;
        error = "";
        try {
            users = await userService.getUsers();
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    }

    function openAddModal() {
        editingUser = null;
        formName = "";
        formEmail = "";
        formPassword = "";
        formRole = "mahasiswa";
        formError = "";
        showModal = true;
    }

    function openEditModal(user) {
        editingUser = user;
        formName = user.name;
        formEmail = user.email;
        formPassword = "";
        formRole = user.role;
        formError = "";
        showModal = true;
    }

    function closeModal() {
        showModal = false;
        editingUser = null;
    }

    async function handleSubmit() {
        if (!formName || !formEmail) {
            formError = "Name and email are required";
            return;
        }

        if (!editingUser && !formPassword) {
            formError = "Password is required for new users";
            return;
        }

        if (formPassword && formPassword.length < 6) {
            formError = "Password must be at least 6 characters";
            return;
        }

        formLoading = true;
        formError = "";

        try {
            const userData = {
                name: formName,
                email: formEmail,
                role: formRole,
            };

            if (formPassword) {
                userData.password = formPassword;
            }

            if (editingUser) {
                await userService.updateUser(editingUser.id, userData);
            } else {
                userData.password = formPassword;
                await userService.createUser(userData);
            }

            closeModal();
            await loadUsers();
        } catch (err) {
            formError = err.message;
        } finally {
            formLoading = false;
        }
    }

    async function handleDelete(user) {
        if (user.id === currentUser?.id) {
            error = "Cannot delete your own account";
            return;
        }

        deleteConfirm = user;
    }

    async function confirmDelete() {
        if (!deleteConfirm) return;

        try {
            await userService.deleteUser(deleteConfirm.id);
            deleteConfirm = null;
            await loadUsers();
        } catch (err) {
            error = err.message;
            deleteConfirm = null;
        }
    }

    function formatDate(dateString) {
        return new Date(dateString).toLocaleDateString("id-ID", {
            year: "numeric",
            month: "short",
            day: "numeric",
        });
    }
</script>

<div class="space-y-6">
    <!-- Header -->
    <div
        class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4"
    >
        <div>
            <h1 class="text-2xl font-bold text-slate-900 dark:text-white">
                User Management
            </h1>
            <p class="text-slate-500 dark:text-slate-400 mt-1">
                Manage system users and their roles
            </p>
        </div>
        <button
            on:click={openAddModal}
            class="inline-flex items-center gap-2 px-4 py-2.5 bg-primary hover:bg-primary/90 text-white font-medium rounded-lg transition-all shadow-lg shadow-primary/25"
        >
            <span class="material-symbols-outlined text-xl">person_add</span>
            Add User
        </button>
    </div>

    <!-- Error Message -->
    {#if error}
        <div
            class="p-4 bg-red-100 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg"
        >
            <p class="text-red-600 dark:text-red-400 flex items-center gap-2">
                <span class="material-symbols-outlined">error</span>
                {error}
            </p>
        </div>
    {/if}

    <!-- Users Table -->
    <div
        class="bg-white dark:bg-[#192633] border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden shadow-sm"
    >
        {#if loading}
            <div class="p-12 text-center">
                <div class="inline-flex items-center gap-3 text-slate-500">
                    <svg class="animate-spin h-6 w-6" viewBox="0 0 24 24">
                        <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                            fill="none"
                        ></circle>
                        <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                    </svg>
                    Loading users...
                </div>
            </div>
        {:else if users.length === 0}
            <div class="p-12 text-center">
                <span
                    class="material-symbols-outlined text-6xl text-slate-300 dark:text-slate-600"
                    >group_off</span
                >
                <p class="mt-4 text-slate-500 dark:text-slate-400">
                    No users found
                </p>
            </div>
        {:else}
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead class="bg-slate-50 dark:bg-slate-800/50">
                        <tr>
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >User</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Email</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Role</th
                            >
                            <th
                                class="px-6 py-4 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Created</th
                            >
                            <th
                                class="px-6 py-4 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider"
                                >Actions</th
                            >
                        </tr>
                    </thead>
                    <tbody
                        class="divide-y divide-slate-100 dark:divide-slate-700"
                    >
                        {#each users as user}
                            <tr
                                class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors"
                            >
                                <td class="px-6 py-4 whitespace-nowrap">
                                    <div class="flex items-center gap-3">
                                        <div
                                            class="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-blue-600 flex items-center justify-center text-white font-bold text-sm"
                                        >
                                            {user.name.charAt(0).toUpperCase()}
                                        </div>
                                        <div>
                                            <p
                                                class="font-medium text-slate-900 dark:text-white"
                                            >
                                                {user.name}
                                            </p>
                                            {#if user.id === currentUser?.id}
                                                <span
                                                    class="text-xs text-primary"
                                                    >(You)</span
                                                >
                                            {/if}
                                        </div>
                                    </div>
                                </td>
                                <td
                                    class="px-6 py-4 whitespace-nowrap text-slate-600 dark:text-slate-300"
                                >
                                    {user.email}
                                </td>
                                <td class="px-6 py-4 whitespace-nowrap">
                                    <span
                                        class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium {user.role ===
                                        'admin'
                                            ? 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400'
                                            : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300'}"
                                    >
                                        {#if user.role === "admin"}
                                            <span
                                                class="material-symbols-outlined text-sm mr-1"
                                                >shield</span
                                            >
                                        {/if}
                                        {user.role}
                                    </span>
                                </td>
                                <td
                                    class="px-6 py-4 whitespace-nowrap text-sm text-slate-500 dark:text-slate-400"
                                >
                                    {formatDate(user.created_at)}
                                </td>
                                <td
                                    class="px-6 py-4 whitespace-nowrap text-right"
                                >
                                    <div
                                        class="flex items-center justify-end gap-2"
                                    >
                                        <button
                                            on:click={() => openEditModal(user)}
                                            class="p-2 text-slate-400 hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                                            title="Edit"
                                        >
                                            <span
                                                class="material-symbols-outlined text-xl"
                                                >edit</span
                                            >
                                        </button>
                                        {#if user.id !== currentUser?.id}
                                            <button
                                                on:click={() =>
                                                    handleDelete(user)}
                                                class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                                                title="Delete"
                                            >
                                                <span
                                                    class="material-symbols-outlined text-xl"
                                                    >delete</span
                                                >
                                            </button>
                                        {/if}
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>

<!-- Add/Edit Modal -->
{#if showModal}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
        <div
            class="bg-white dark:bg-[#192633] border border-slate-200 dark:border-slate-700 rounded-2xl shadow-2xl w-full max-w-md"
        >
            <div
                class="px-6 py-4 border-b border-slate-200 dark:border-slate-700"
            >
                <h2 class="text-xl font-bold text-slate-900 dark:text-white">
                    {editingUser ? "Edit User" : "Add New User"}
                </h2>
            </div>

            <form on:submit|preventDefault={handleSubmit} class="p-6 space-y-4">
                {#if formError}
                    <div
                        class="p-3 bg-red-100 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg"
                    >
                        <p class="text-red-600 dark:text-red-400 text-sm">
                            {formError}
                        </p>
                    </div>
                {/if}

                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
                        >Name</label
                    >
                    <input
                        type="text"
                        bind:value={formName}
                        class="w-full h-11 px-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50"
                        placeholder="Enter name"
                        required
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
                        >Email</label
                    >
                    <input
                        type="email"
                        bind:value={formEmail}
                        class="w-full h-11 px-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50"
                        placeholder="Enter email"
                        required
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
                    >
                        Password {editingUser
                            ? "(leave empty to keep current)"
                            : ""}
                    </label>
                    <input
                        type="password"
                        bind:value={formPassword}
                        class="w-full h-11 px-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50"
                        placeholder={editingUser
                            ? "Leave empty to keep current"
                            : "Enter password"}
                    />
                </div>

                <div>
                    <label
                        class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
                        >Role</label
                    >
                    <select
                        bind:value={formRole}
                        class="w-full h-11 px-4 bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50"
                    >
                        <option value="mahasiswa">Mahasiswa</option>
                        <option value="admin">Admin</option>
                    </select>
                </div>

                <div class="flex gap-3 pt-4">
                    <button
                        type="button"
                        on:click={closeModal}
                        class="flex-1 h-11 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 font-medium rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        type="submit"
                        disabled={formLoading}
                        class="flex-1 h-11 bg-primary hover:bg-primary/90 disabled:bg-primary/50 text-white font-medium rounded-lg transition-colors flex items-center justify-center gap-2"
                    >
                        {#if formLoading}
                            <svg
                                class="animate-spin h-5 w-5"
                                viewBox="0 0 24 24"
                            >
                                <circle
                                    class="opacity-25"
                                    cx="12"
                                    cy="12"
                                    r="10"
                                    stroke="currentColor"
                                    stroke-width="4"
                                    fill="none"
                                ></circle>
                                <path
                                    class="opacity-75"
                                    fill="currentColor"
                                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                                ></path>
                            </svg>
                        {:else}
                            {editingUser ? "Save Changes" : "Create User"}
                        {/if}
                    </button>
                </div>
            </form>
        </div>
    </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if deleteConfirm}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
    >
        <div
            class="bg-white dark:bg-[#192633] border border-slate-200 dark:border-slate-700 rounded-2xl shadow-2xl w-full max-w-sm p-6"
        >
            <div class="text-center">
                <div
                    class="w-16 h-16 mx-auto mb-4 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center"
                >
                    <span
                        class="material-symbols-outlined text-4xl text-red-500"
                        >warning</span
                    >
                </div>
                <h3
                    class="text-xl font-bold text-slate-900 dark:text-white mb-2"
                >
                    Delete User?
                </h3>
                <p class="text-slate-500 dark:text-slate-400 mb-6">
                    Are you sure you want to delete <strong
                        class="text-slate-900 dark:text-white"
                        >{deleteConfirm.name}</strong
                    >? This action cannot be undone.
                </p>
                <div class="flex gap-3">
                    <button
                        on:click={() => (deleteConfirm = null)}
                        class="flex-1 h-11 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 font-medium rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        on:click={confirmDelete}
                        class="flex-1 h-11 bg-red-500 hover:bg-red-600 text-white font-medium rounded-lg transition-colors"
                    >
                        Delete
                    </button>
                </div>
            </div>
        </div>
    </div>
{/if}
