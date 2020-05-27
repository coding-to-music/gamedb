const $appsComparePage = $('#apps-compare-page');

if ($appsComparePage.length > 0) {

    loadAppPlayersChart();
    loadGroupChart();

    function loadAppPlayersChart() {

        const defaultAppChartOptions = {
            chart: {
                type: 'spline',
                backgroundColor: 'rgba(0,0,0,0)',
            },
            title: {
                text: ''
            },
            subtitle: {
                text: ''
            },
            credits: {
                enabled: false
            },
            legend: {
                enabled: true
            },
            xAxis: {
                title: {text: ''},
                type: 'datetime',
            },
            yAxis: {
                allowDecimals: false,
                title: {text: ''},
                min: 0,
                opposite: false,
                labels: {
                    formatter: function () {
                        return this.value.toLocaleString();
                    },
                },
                visible: true,
            },
            plotOptions: {
                series: {
                    marker: {
                        enabled: false
                    },
                }
            },
            colors: ['#007bff', '#28a745', '#e83e8c', '#ffc107', '#343a40'],
            tooltip: {
                formatter: function () {
                    return this.y.toLocaleString() + ' players on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                },
            },
        };

        $.ajax({
            type: "GET",
            url: '/games/' + $appsComparePage.attr('data-id') + '/players.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['max_player_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('players-chart', $.extend(true, {}, defaultAppChartOptions, {
                    series: series,
                }));

            },
        });

        $.ajax({
            type: "GET",
            url: '/games/' + $appsComparePage.attr('data-id') + '/players2.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: appNames[datum.key],
                        data: datum['value']['max_player_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('players-chart2', $.extend(true, {}, defaultAppChartOptions, {
                    series: series,
                }));

            },
        });
    }

    function loadGroupChart($page = null) {

        $.ajax({
            type: "GET",
            url: '/groups/' + $appsComparePage.attr('data-group-id') + '/members.json',
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                let series = [];

                for (const datum of data) {
                    series.push({
                        name: groupNames[datum.key],
                        data: datum['value']['max_members_count'],
                        connectNulls: true,
                    });
                }

                Highcharts.chart('group-chart', {
                    chart: {
                        type: 'spline',
                        backgroundColor: 'rgba(0,0,0,0)',
                    },
                    title: {
                        text: ''
                    },
                    subtitle: {
                        text: ''
                    },
                    credits: {
                        enabled: false,
                    },
                    legend: {
                        enabled: true,
                    },
                    xAxis: {
                        title: {
                            text: ''
                        },
                        type: 'datetime'

                    },
                    yAxis: {
                        allowDecimals: false,
                        title: {
                            text: ''
                        },
                        labels: {
                            formatter: function () {
                                return this.value.toLocaleString();
                            },
                        },
                    },
                    tooltip: {
                        formatter: function () {
                            return this.y.toLocaleString() + ' members on ' + moment(this.key).format("dddd DD MMM YYYY @ HH:mm");
                        },
                    },
                    series: series,
                });
            },
        });
    }
}