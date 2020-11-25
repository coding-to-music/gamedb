const $appsAchievementsComparePage = $('#apps-achievements-compare-page');

if ($appsAchievementsComparePage.length > 0) {

    loadFriends(function ($chosen) {

        const val = $chosen.val();
        if (val) {

            let pieces = window.location.pathname.split('/');
            let ids = pieces.length === 5 ? pieces[4].split(',') : [];

            ids.push(val);
            ids = [...new Set(ids)]; // Unique

            window.location.href = '/games/' + $appsAchievementsComparePage.attr('data-app-id')
                + '/compare-achievements/' + ids.join(',');
        }
    });
}
