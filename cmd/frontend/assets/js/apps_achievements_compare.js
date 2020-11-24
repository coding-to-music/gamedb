const $appsAchievementsComparePage = $('#apps-achievements-compare-page');

if ($appsAchievementsComparePage.length > 0) {

    loadFriends(function ($chosen) {

        const val = $chosen.val();
        if (val) {

            let pieces = window.location.pathname.substring(window.location.pathname.lastIndexOf('/') + 1);
            let playerIds = pieces.split(',')
            playerIds.push(val);

            pieces = [...new Set(playerIds)]; // Unique

            window.location.href = '/games/' + $appsAchievementsComparePage.attr('data-app-id')
                + '/compare-achievements/' + pieces.join(',');
        }
    });
}
