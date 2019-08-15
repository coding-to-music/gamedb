function highLightOwnedGames() {
    if (user.isLoggedIn) {

        let games = localStorage.getItem('games');
        if (games != null) {
            games = JSON.parse(games);
            if (games != null) {
                $('[data-app-id]').each(function () {
                    const id = $(this).attr('data-app-id');
                    if (games.indexOf(parseInt(id)) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                });
            }
        }

        let groups = localStorage.getItem('groups');
        if (groups != null) {
            groups = JSON.parse(groups);
            if (groups != null) {
                $('[data-group-id]').each(function () {
                    const id = $(this).attr('data-group-id');
                    if (groups.indexOf(id) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                    const id64 = $(this).attr('data-group-id64');
                    if (groups.indexOf(id64) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                });
            }
        }
    }
}
