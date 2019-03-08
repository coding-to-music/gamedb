if ($('#home-page').length > 0) {

    $.ajax({
        type: "GET",
        url: '/home/charts.json',
        dataType: 'json',
        success: function (datas, textStatus, jqXHR) {

            if (datas === null) {
                return
            }

            $('div[data-app-id]').each(function (index) {

                let data = {};
                const appID = $(this).attr('data-app-id');

                if (datas !== null && appID in datas && 'max_player_count' in datas[appID]) {
                    data = datas[appID]['max_player_count'];
                } else {
                    data = [];
                }

                Highcharts.chart(this, {
                    chart: {
                        type: 'area',
                        margin: [0, 0, 0, 0],
                        skipClone: true,
                        backgroundColor: null,
                        height: 32,
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
                        enabled: false
                    },
                    xAxis: {
                        title: {text: null},
                        labels: {enabled: false},
                        type: 'datetime',
                    },
                    yAxis: {
                        title: {text: null},
                        labels: {enabled: false},
                        min: 0,
                    },
                    tooltip: {
                        hideDelay: 0,
                        outside: true,
                        shared: true,
                        formatter: function () {
                            return this.y.toLocaleString() + ' players on ' + moment(this.x).format("DD MMM YYYY @ HH:mm");
                        },
                        style: {
                            'width': '500px',
                        }
                    },
                    series: [
                        {
                            color: '#28a745',
                            data: data,
                        },
                    ],
                });

            });

        },
    });

}
