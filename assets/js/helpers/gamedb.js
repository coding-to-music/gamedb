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
    }
}