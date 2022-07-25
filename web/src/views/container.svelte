<script lang="ts">
    import { onMount } from "svelte";

    import { filedrop } from "filedrop-svelte";
	import type { Files, FileDropOptions } from "filedrop-svelte";
	let options: FileDropOptions = {};
	let files: Files;

    let loading = true;

    let newContainer = false;

    export let containerKey;
    let containerInfo = {
        empty: true,
        key: '',
        filename: '',
        previewGenerated: false,
    };

    let accessDenied = false;

    let previewCacheBust = 'default';

    const urlParams = new URLSearchParams(window.location.search);
    
    let signature = urlParams.get('s');
    let expiration = urlParams.get('e');

    async function getContainerInfo() {
        const res = await fetch(`/api/containers/${containerKey}`, {
            headers: {
                'X-Uber-Signature': signature,
                'X-Uber-Signature-Expire': expiration
            }
        });

        if (res.status == 401) {
            accessDenied = true;
        }

        if (res.status === 200) {
            containerInfo = await res.json();

            console.log(containerInfo);
        }

        if (res.status == 404) {
            newContainer = true;
        }

        loading = false;
    }

    async function deleteFile () {
        const res = await fetch(`/api/files/${containerKey}`,
        {
            method: "DELETE",
            headers: {
                'X-Uber-Signature': signature,
                'X-Uber-Signature-Expire': expiration
            },
        });

        if (res.status == 401) {
            accessDenied = true;
        }

        getContainerInfo();
    }

    async function confirmDelete () {
        var confirmed = confirm("Are you sure you want to delete this file?");
        if (confirmed) {
            deleteFile();
        }
    };

    function uploadProgress (e) {
        console.log(e);
    }

    function uploadComplete (e) {
        console.log(e);
    }

    function uploadFailed (e) {
        console.log(e);
    }

    function uploadFile (file) {
        loading = true;
        var formData = new FormData();

        formData.append("container_key", `/${containerKey}`);

        formData.append("file", file);

        var xhr = new XMLHttpRequest();

        xhr.addEventListener('progress', uploadProgress, false);
        xhr.addEventListener('load', uploadComplete, false);
        xhr.addEventListener('error', uploadFailed, false);

        var postUrl;
        if (newContainer) {
            postUrl = '/api/containers';
        } else {
            postUrl = `/api/containers/${containerKey}`;
        }

        xhr.open('POST', postUrl, true);
        xhr.setRequestHeader('Filename', file.name);
        xhr.setRequestHeader('X-Uber-Signature', signature);
        xhr.setRequestHeader('X-Uber-Signature-Expire', expiration);

        xhr.onload = function () {
            var body = JSON.parse(xhr.responseText);

            previewCacheBust = new Date().getTime().toString();
            
            getContainerInfo();

            loading = false;
        };

        xhr.send(formData);
    }

    getContainerInfo();

    async function getDownloadUrl() {
        const res = await fetch(`/api/files/${containerKey}?r=true`, {
            headers: {
                'X-Uber-Signature': signature,
                'X-Uber-Signature-Expire': expiration
            }
        });

        if (res.status == 401) {
            accessDenied = true;
        }

        if (res.status === 200) {
            const result = await res.json();

            console.log(result);

            downloadFile(result.downloadLink)
        }

        loading = false;
    }

    function downloadFile(link) {
        window.open(link, '_blank');
    }

    function drop(e) {
        e.preventDefault();
        e.stopPropagation();
        console.log('dropped file!', e.detail.files);

        uploadFile(e.detail.files.accepted[0]);
    }

</script>

<div style="height:100%;">
        {#if loading}
		    <img src="/images/ajax-loader.gif" />
        {/if}

        {#if accessDenied}
        <div class="accessDenied">
            <h3 style="color:red;">Access Denied<h3>

            <div>Invalid or expired signature.</div>
            <div>If embeded coming back to this page might be necessary</div>
        </div>
        {:else}
        <div class="dropzone" use:filedrop={options} on:filedrop={drop}>
            {#if containerInfo.empty}
            <div class="emptyContainer">
                <h3>Empty Container<h3>

                <div>Drag File or Click to Upload</div>
            </div>
            {:else}
                <div>Filename: {containerInfo.filename}</div>

                {#if containerInfo.previewGenerated}
                    <img class="preview-image" src="/api/previews{containerInfo.key}?i={previewCacheBust}" style="text-align: center;margin: 0 auto;display: table-cell;max-width:100px;cursor:pointer;" />
                {:else}
                    <div ng-if="errorLoading" style="text-align:center;">Unable to preview</div>
                {/if}
            {/if}
        </div>
        {/if}

        {#if !containerInfo.empty}
        <div class="btn-group">
            <button type="button" class="btn btn-default pull-left" on:click={getDownloadUrl}>Download</button>
            <button type="button" class="btn btn-default" on:click={confirmDelete}>Delete</button>
        </div>
        {/if}

		<input class="fileInput" type="file" style="display:none;" />
</div>

<style>
    .dropzone {

    }

    .dropzone:focus {
        border-style: dashed;
        border-color: #2196f3;
        border-width: 2px;
        border-radius: 2px;
    }
</style>