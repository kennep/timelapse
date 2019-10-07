var idToken;

function init() {
    gapi.load('auth2', function() {
        gapi.auth2.init({
            'client_id': '91541969634-dpoa610se2n1guhmsbrqmpvbpoucdntu.apps.googleusercontent.com'
        }).then(function(auth2) {
            //auth2.isSignedIn.listen(onSignIn);
            //auth2.currentUser.listen(onSignIn);
            
            if (auth2.isSignedIn.get() == true) {
                console.log("User is signed in");
                //auth2.signIn();
                onSignIn(auth2.currentUser.get());
            } else {
                console.log("User is not signed in");
                signedOut();
            }
        });
    });
}

function signedOut()
{
    doneLoading();
    $("#welcome").show();
    $("#signedinbox").hide();
    idToken = null;
    gapi.signin2.render('google-signin',
                        {
                            'onsuccess': onSignIn
                        });
}


async function onSignIn(googleUser) {
    idToken = googleUser.getAuthResponse().id_token;
    
    var profile = googleUser.getBasicProfile();
    if(!profile) return;

    
    console.log('ID: ' + profile.getId()); // Do not send to your backend! Use an ID token instead.
    console.log('Name: ' + profile.getName());
    console.log('Image URL: ' + profile.getImageUrl());
    console.log('Email: ' + profile.getEmail()); // This is null if the 'email' scope is not present.
    $("#welcome").hide();
    if(profile.getImageUrl()) {
        $("#profileimage").attr("src", profile.getImageUrl());
    } else {
        $("#profileimage").hide();
    }
    $("#profilename").text(profile.getName());
    $("#signedinbox").show();


    var projects = await getProjects();
    var entries = await getEntries();
    

    startProjects(projects);
    $("#entries").text(JSON.stringify(entries));

}

async function getProjects()
{
    return jsonRequest("/projects");
}

async function getEntries()
{
    return jsonRequest("/entries");
}

async function jsonRequest(path)
{
    loading();
    var headers = new Headers();
    headers.append('Authorization', 'Bearer ' + idToken);

    var response = await fetch(path, {'headers': headers});

    doneLoading();

    var body = 'Server sent a malformed response';
    var ct = response.headers.get("Content-Type");
    if(ct != null && ct.includes("json")) {
        body = await response.json();
    } else {
        body = {'message': await response.text()};
    }
    
    if(!response.ok) {        
        showServerError(body['message'] ? body['message'] : JSON.stringify(body));
        return {'error': body};
    }
    return body;
}

function signOut() {
    var auth2 = gapi.auth2.getAuthInstance();
    auth2.signOut().then(function () {
        signedOut();
        console.log('User signed out.');
    });
}

function chooseAccount() {
    console.log('Choose account');
    var auth2 = gapi.auth2.getAuthInstance();
    auth2.signIn({'prompt': 'select_account'}).then(onSignIn);
}

function showServerError(errorText)
{
    $("#servererrormessage").text(errorText);
    $("#servererror").show();
}

function loading()
{
    $("#loading").show();
}

function doneLoading()
{
    $("#loading").hide();
}

function startProjects(projects)
{
    $("#startentry").show();
    projects.forEach(function(project) {
        var div = $('<span> </span>');
        div.appendTo('#startbuttons');
        $('<button type="button" class="btn btn-outline-primary"></button>').text(project['name'] || project['description'] || 'Unnamed project').appendTo(div);
    });
}

if(window.gapi) {
    init()
}
