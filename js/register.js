function onSubmit(token) {
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmpassword').value;

    console.log('doing things');

    console.log('doing things');
    if (password != confirmPassword) {
        console.log('mismatching passwords');
        return;
    }
    console.log('doing things');

    document.getElementById('register-form').submit();
}
