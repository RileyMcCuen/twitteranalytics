document.getElementById('search').addEventListener('click', async (ev) => {
    const resp = await fetch(
        `http://localhost/api/analysis?name=${
            document.getElementById('twitter-handle').value
        }`
    );
    if (resp.ok) {
        alert('Good Response');
    } else {
        alert('BAD RESPONSE: ' + resp.status + ': ' + (await resp.text()));
    }
});
