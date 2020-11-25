const $appsAchievementsComparePage = $('#apps-achievements-compare-page');

if ($appsAchievementsComparePage.length > 0) {

    const appID = $appsAchievementsComparePage.attr('data-app-id');

    loadFriends(appID, false, function ($chosen) {

        const val = $chosen.val();
        if (val) {

            let pieces = window.location.pathname.split('/');
            let ids = pieces.length === 5 ? pieces[4].split(',') : [];

            ids.push(val);
            ids = [...new Set(ids)]; // Unique

            window.location.href = '/games/' + appID + '/compare-achievements/' + ids.join(',');
        }
    });
}
