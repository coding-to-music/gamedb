function highLightOwnedGames($element) {

    if (!$element) {
        $element = $('body');
    }

    if (user.isLoggedIn) {

        let games = localStorage.getItem('gamedb-games');
        if (games != null) {
            games = JSON.parse(games);
            if (games != null) {
                $element.find('[data-app-id]').each(function () {
                    const id = $(this).attr('data-app-id');
                    if (games.indexOf(parseInt(id)) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                });
            }
        }

        let groups = localStorage.getItem('gamedb-groups');
        if (groups != null) {
            groups = JSON.parse(groups);
            if (groups != null) {
                $element.find('[data-group-id]').each(function () {
                    const id = $(this).attr('data-group-id');
                    if (groups.indexOf(id) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                });
            }
        }

        let badges = localStorage.getItem('gamedb-badges');
        if (badges != null) {
            badges = JSON.parse(badges);
            if (badges != null) {
                $element.find('[data-badge-id]').each(function () {
                    const id = $(this).attr('data-badge-id');
                    if (badges.indexOf(parseInt(id)) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                });
            }
        }
    }
}
